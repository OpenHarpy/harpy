package main

import (
	"client-engine/config"
	"client-engine/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	logger.SetupLogging()
	logger.Info("Starting Client Engine", "MAIN")
	lm := NewLiveMemory()

	// Create wait groups for the servers
	var exitMainServerChan = make(chan bool)
	var exitCallbackServerChan = make(chan bool)
	var wg sync.WaitGroup
	wg.Add(2)

	// Start the gRPC server

	port := config.GetConfigs().GetConfigsWithDefault("port", "50051")
	err := NewCEServer(exitMainServerChan, &wg, lm, port)
	if err != nil {
		logger.Error("Failed to start gRPC server", "MAIN", err)
		return
	}

	// Start the gRPC callback server
	port = config.GetConfigs().GetConfigsWithDefault("callback.port", "50052")
	err = StartCallbackServer(exitCallbackServerChan, &wg, lm, port)
	if err != nil {
		logger.Error("Failed to start gRPC callback server", "MAIN", err)
		return
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
