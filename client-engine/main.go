package main

import (
	"client-engine/config"
	"client-engine/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Required Configs for the resource manager
var requiredConfigs = []string{
	"harpy.clientEngine.grpcServer.servePort",
	"harpy.clientEngine.grpcServer.serveHost",
	"harpy.clientEngine.grpcCallbackServer.serveHost",
	"harpy.clientEngine.grpcCallbackServer.servePort",
	"harpy.clientEngine.resourceManager.uri",
}

func main() {
	logger.SetupLogging()
	logger.Info("Starting Client Engine", "MAIN")
	lm := NewLiveMemory()
	// Validate the required configs
	err := config.GetConfigs().ValitateRequiredConfigs(requiredConfigs)
	if err != nil {
		logger.Error("Failed to validate required configs", "MAIN", err)
		os.Exit(1)
	}

	// Create wait groups for the servers
	var exitMainServerChan = make(chan bool)
	var exitCallbackServerChan = make(chan bool)
	var wg sync.WaitGroup

	// Start the gRPC server
	port := config.GetConfigs().GetConfigWithDefault("harpy.clientEngine.grpcServer.servePort", "50051")
	err = NewCEServer(exitMainServerChan, &wg, lm, port)
	if err != nil {
		logger.Error("Failed to start gRPC server", "MAIN", err)
		os.Exit(1)
	}

	// Start the gRPC callback server
	port = config.GetConfigs().GetConfigWithDefault("harpy.clientEngine.grpcCallbackServer.servePort", "50052")
	err = StartCallbackServer(exitCallbackServerChan, &wg, lm, port)
	if err != nil {
		logger.Error("Failed to start gRPC callback server", "MAIN", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received signal to shutdown", "MAIN")
		exitMainServerChan <- true
		exitCallbackServerChan <- true
	}()
	// Wait for the server to exit
	wg.Wait()
	logger.Info("Client Engine stopped", "MAIN")
}
