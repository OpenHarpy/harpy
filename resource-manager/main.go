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

// Required Configs for the resource manager
var requiredConfigs = []string{
	"harpy.resourceManager.grpcServer.servePort",
	"harpy.resourceManager.grpcServer.serveHost",
	"harpy.resourceManager.httpServer.servePort",
	"harpy.resourceManager.httpServer.serveHost",
	"harpy.resourceManager.nodeProvider",
	"harpy.resourceManager.nodeCatalog.location",
	"harpy.resourceManager.database.provider",
	"harpy.resourceManager.database.uri",
}

func NodeLoader() error {
	// Load the node catalog
	catalogData := config.GetConfigs().GetConfigWithDefault("harpy.resourceManager.nodeCatalog.location", "catalog.json")
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
	nodeProvider := config.GetConfigs().GetConfigWithDefault("harpy.resourceManager.nodeProvider", "local")
	var provider providers.ProviderInterface
	if nodeProvider == "local" {
		command, ok := config.GetConfigs().GetConfig("harpy.resourceManager.localProvider.command")
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

func main() {
	logger.SetupLogging()
	logger.Info("Starting Resource Manager", "MAIN")

	// Validate the required configs
	err := config.GetConfigs().ValitateRequiredConfigs(requiredConfigs)
	if err != nil {
		logger.Error("Failed to validate required configs", "MAIN", err)
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
	port := config.GetConfigs().GetConfigWithDefault("harpy.resourceManager.grpcServer.servePort", "50050")
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
	httpPort := config.GetConfigs().GetConfigWithDefault("harpy.resourceManager.httpServer.servePort", "8080")
	staticFiles := config.GetConfigs().GetConfigWithDefault("harpy.resourceManager.ui.staticFiles", "")
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
