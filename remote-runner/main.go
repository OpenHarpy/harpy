package main

import (
	"os"
	"os/signal"
	"remote-runner/config"
	"remote-runner/logger"
	"syscall"

	"github.com/sirupsen/logrus"
)

// Begin LiveMemory
type LiveMemory struct {
	// The Live Memory is a struct that holds all the live data of the server
	// It is used to store all the live data of the server
	Process                map[string]*Process
	Callback               map[string]*Callback
	CallbackClients        map[string]*CallbackClient
	NodeStatusUpdateClient *NodeStatusUpdateClient
}

func main() {
	// Setup logging
	logger.SetupLogging()

	// Argument parsing
	// This is where we will parse the arguments
	// We will need to parse the arguments to get the nodeID and the resource manager's address

	// By default if we try to access an index that does not exist in the array we will get an error
	// We will need to handle this error
	if len(os.Args) < 4 {
		logger.Error("Not enough arguments passed - nodeID and resource manager address are required", "MAIN", nil)
		os.Exit(1) // Exit code 1 is for general errors in the program
		return
	}
	nodeID := os.Args[1]
	nodeType := os.Args[2]
	resourceManagerAddress := os.Args[3]
	port := config.GetConfigs().GetConfigsWithDefault("port", "50053")

	logger.Info("Node ID", "MAIN", logrus.Fields{"nodeID": nodeID})
	logger.Info("Node Type", "MAIN", logrus.Fields{"nodeType": nodeType})
	logger.Info("Resource Manager Address", "MAIN", logrus.Fields{"resourceManagerAddress": resourceManagerAddress})

	// Create the Node Reporter client
	nodeStatusUpdateClient := NewNodeStatusUpdateClient(
		nodeID, nodeType, "localhost:"+port, resourceManagerAddress,
	)
	nodeStatusUpdateClient.connect()
	nodeStatusUpdateClient.SetNodeBooting()

	// We then need to inform the resource manager that we are being setup

	logger.Info("Starting Remote Runner", "MAIN")
	NodeSetup()
	// Here we should send a message to the node router to let it know that we are ready to accept commands
	logger.Info("Node setup complete", "MAIN")

	// Make the live memory struct
	lm := &LiveMemory{
		Process:                make(map[string]*Process),
		Callback:               make(map[string]*Callback),
		CallbackClients:        make(map[string]*CallbackClient),
		NodeStatusUpdateClient: nodeStatusUpdateClient,
	}

	// Start the gRPC server
	exitMainServer := make(chan bool)
	panicMainServer := make(chan bool)
	waitServerExitChan := make(chan bool)
	go NewNodeServer(exitMainServer, panicMainServer, waitServerExitChan, port, lm)

	// Start the event loop
	exitEventLoop := make(chan bool)
	waitEventLoopExitChan := make(chan bool)
	go ProcessEventLoop(exitEventLoop, waitEventLoopExitChan, lm)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	nodeStatusUpdateClient.SetNodeReady()

	go func() {
		<-sigChan
		logger.Info("Received signal to shutdown", "MAIN")
		nodeStatusUpdateClient.SetNodeShuttingDown()
		exitEventLoop <- true
		exitMainServer <- true
	}()

	// Wait for the server to exit
	<-waitServerExitChan
	<-waitEventLoopExitChan
	logger.Info("Exiting Remote Runner", "MAIN")
	//ExitCleanAll(lm)
	// TODO: Implement ExitCleanAll
	nodeStatusUpdateClient.SetNodeShutdown()
	nodeStatusUpdateClient.disconnect()
	logger.Info("Node shutdown complete", "MAIN")
}