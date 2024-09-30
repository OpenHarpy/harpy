package main

import (
	"client-engine/config"
	"client-engine/logger"
	"client-engine/task"
	"os"
	"os/signal"
	"syscall"
)

// Begin LiveMemory
type LiveMemory struct {
	// The Live Memory is a struct that holds all the live data of the server
	// It is used to store all the live data of the server
	Sessions           map[string]*task.Session
	TaskSetDefinitions map[string]*task.TaskSet
	TaskDefinitions    map[string]*TaskDefinition
}

func main() {
	logger.SetupLogging()
	logger.Info("Starting Client Engine", "MAIN")
	lm := &LiveMemory{
		Sessions:           make(map[string]*task.Session),
		TaskSetDefinitions: make(map[string]*task.TaskSet),
		TaskDefinitions:    make(map[string]*TaskDefinition),
	}

	// Start the gRPC server
	exitMainServer := make(chan bool)
	panicMainServer := make(chan bool)
	port := config.GetConfigs().GetConfigsWithDefault("port", "50051")
	go NewCEServer(exitMainServer, panicMainServer, port, lm)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received signal to shutdown", "MAIN")
		exitMainServer <- true
	}()

	// Handle panics
	go func() {
		<-panicMainServer
		logger.Error("Panic in main server", "MAIN", nil)
	}()

	// Wait for the server to exit
	<-exitMainServer
	logger.Info("Client Engine stopped", "MAIN")
}
