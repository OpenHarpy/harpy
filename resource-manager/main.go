package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"resource-manager/config"
	"resource-manager/logger"
	"resource-manager/objects"
	"resource-manager/providers"
	lp "resource-manager/providers/local-provider"
	"sync"
	"syscall"
)

func NodeLoader() error {
	// Load the node catalog
	catalogData := config.GetConfigs().GetConfigsWithDefault("node_catalog", "catalog.json")
	// Read the json file
	jsonFile, err := os.ReadFile(catalogData)
	if err != nil {
		logger.Error("Failed to open node catalog", "MAIN", err)
		return err
	}

	// Parse the json file
	var nodes []objects.NodeCatalog
	err = json.Unmarshal(jsonFile, &nodes)
	if err != nil {
		logger.Error("Failed to parse node catalog", "MAIN", err)
		return err
	}

	// Set the default configurations For each key value pair in the default configurations
	for _, node := range nodes {
		// We add the key value pair on the database
		node.Sync()
	}
	return nil
}

func GetProvider() (providers.ProviderInterface, error) {
	// Load the node catalog
	err := NodeLoader()
	if err != nil {
		return nil, err
	}
	nodeProvider := config.GetConfigs().GetConfigsWithDefault("node-provider", "local")
	var provider providers.ProviderInterface
	if nodeProvider == "local" {
		command, ok := config.GetConfigs().GetConfig("local-provider-command")
		if !ok {
			return nil, errors.New("local provider command not set, cannot continue")
		}
		provider = lp.NewLocalProvider(command)
	} else {
		return nil, errors.New("invalid provider option was set")
	}
	err = providers.StartProvider(provider)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

// Load the default environment configurations
func LoadDefaultConfigs() error {
	confLocation := config.GetConfigs().GetConfigsWithDefault("default_env_configs", "default_env_configs.json")
	// Read the json file
	jsonFile, err := os.ReadFile(confLocation)
	if err != nil {
		logger.Error("Failed to open default environment configurations", "MAIN", err)
		return err
	}

	// Parse the json file
	var configs map[string]string
	err = json.Unmarshal(jsonFile, &configs)
	if err != nil {
		logger.Error("Failed to parse default environment configurations", "MAIN", err)
		return err
	}

	// Set the default configurations For each key value pair in the default configurations
	for key, value := range configs {
		// We add the key value pair on the database
		conf := &objects.Config{
			ConfigKey:   key,
			ConfigValue: value,
		}
		conf.Sync()
	}
	return nil
}

func main() {
	logger.SetupLogging()
	logger.Info("Starting Resource Manager", "MAIN")

	// Load the default configurations
	err := LoadDefaultConfigs()
	if err != nil {
		logger.Error("Failed to load default configurations", "MAIN", err)
		return
	}

	// Add some local nodes
	runningProvider, err := GetProvider()
	if err != nil {
		logger.Error("Failed to initialize provider", "MAIN", err)
		return
	}

	// Waitgroup for the different components
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Start the gRPC server
	exitMainServer := make(chan bool)
	port := config.GetConfigs().GetConfigsWithDefault("port", "50050")
	err = NewResourceAllocServer(exitMainServer, &wg, port)
	if err != nil {
		logger.Error("Failed to start gRPC server", "MAIN", err)
		return
	}

	// Start the event loop
	exitEventLoop := make(chan bool)
	err = ProcessEventLoop(runningProvider, exitEventLoop, &wg)
	if err != nil {
		logger.Error("Failed to start event loop", "MAIN", err)
		exitEventLoop <- true
		wg.Wait()
		return
	} else {
		wg.Add(1)
	}

	// Start the HTTP server
	exitHttpServer := make(chan bool)
	httpPort := config.GetConfigs().GetConfigsWithDefault("port", "8080")
	staticFiles := config.GetConfigs().GetConfigsWithDefault("ui-static-files", "")
	err = StartServer(exitHttpServer, &wg, httpPort, staticFiles)
	if err != nil {
		logger.Error("Failed to start HTTP server", "MAIN", err)
		exitMainServer <- true
		exitEventLoop <- true
		wg.Wait()
		return
	} else {
		wg.Add(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received signal to shutdown", "MAIN")
		// Cleanup the provider first as it may have resources that need to be cleaned up
		runningProvider.Cleanup()
		exitMainServer <- true
		exitEventLoop <- true
		exitHttpServer <- true
	}()

	// Wait for the server to exit
	wg.Wait()
	// Stop the provider
	logger.Info("Exiting Resource Manager", "MAIN")
}
