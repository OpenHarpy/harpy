package main

import (
	"encoding/json"
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

// Begin LiveMemory
// For debugging purposes we will have a function to add some local running nodes
//
//	This will be changed to PROVIDERS in the future
//	Providers will be the implementation of the node instancing and destruction process
//	Examples of providers are AWS, GCP, Azure, local, etc.
func AddLocalProvider() providers.ProviderInterface {
	NodeCatalog := &objects.NodeCatalog{
		NodeType:         "small-4cpu-8gb",
		NumCores:         4,
		AmountOfMemory:   8,
		AmountOfStorage:  100,
		NodeMaxCapacity:  1, // This is the maximum number of nodes that can be created of this type
		NodeWarmpoolSize: 1,
		NodeIdleTimeout:  60, // This is the time in seconds that a node can be idle before it is destroyed
	}
	NodeCatalog.Sync() // This will sync the node catalog to the database

	// For the local provider we need to clean all the nodes
	// This is because the local provider is used for debugging and testing purposes only
	// We do not want to have any nodes in the database

	// Get all the nodes
	nodes := objects.GetLiveNodes()
	for _, node := range nodes {
		node.Delete()
	}

	// Create a new local provider
	provider := lp.NewLocalProvider()
	return provider
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
	runningProvider := AddLocalProvider()
	err = providers.StartProvider(runningProvider)
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
	err = StartServer(exitHttpServer, &wg)
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
