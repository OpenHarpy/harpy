package main

import (
	"os"
	"os/signal"
	"resource-manager/config"
	"resource-manager/logger"
	"syscall"
)

// Begin LiveMemory
// For debugging purposes we will have a function to add some local running nodes
//
//	This will be changed to PROVIDERS in the future
//	Providers will be the implementation of the node instancing and destruction process
//	Examples of providers are AWS, GCP, Azure, local, etc.
func AddLocalNodes() {
	// TODO: Implement the provider
	NodeCatalog := &NodeCatalog{
		NodeType:         "small-4cpu-8gb",
		NumCores:         4,
		AmountOfMemory:   8,
		AmountOfStorage:  100,
		NodeMaxCapacity:  1, // This is the maximum number of nodes that can be created of this type
		NodeWarmpoolSize: 1,
		NodeIdleTimeout:  60, // This is the time in seconds that a node can be idle before it is destroyed
	}
	NodeCatalog.Sync() // This will sync the node catalog to the database
	// This will make it so that nodes of this type can be requested
	// This is a local node
	liveNode := &LiveNode{
		NodeID:          "local-1",
		NodeType:        "small-4cpu-8gb",
		NodeGRPCAddress: "",
		NodeStatus:      LiveNodeStatusEnum_NODE_STATUS_UNKNOWN,
	}
	// The node will assign itself to the resource manager
	// Upon spinning up the node should "know" its own ID and it should "know" the address of the resource manager
	// This is done by passing the nodeID and the resource manager's address as an argument to the remote-runner's main process
	// As this is just a debug function we will make sure that this is done correctly
	// But in the future this will be done by the provider itself
	liveNode.Sync() // This will sync the live node to the database
}

func main() {
	logger.SetupLogging()
	logger.Info("Starting Resource Manager", "MAIN")

	// Add some local nodes
	AddLocalNodes()

	// Start the gRPC server
	exitMainServer := make(chan bool)
	waitServerExitChan := make(chan bool)
	port := config.GetConfigs().GetConfigsWithDefault("port", "50050")
	go NewResourceAllocServer(exitMainServer, waitServerExitChan, port)

	// Start the event loop
	exitEventLoop := make(chan bool)
	waitEventLoopExitChan := make(chan bool)
	go ProcessEventLoop(exitEventLoop, waitEventLoopExitChan)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received signal to shutdown", "MAIN")
		exitMainServer <- true
		exitEventLoop <- true
	}()

	// Wait for the server to exit
	<-waitServerExitChan
	<-waitEventLoopExitChan
	logger.Info("Exiting Resource Manager", "MAIN")
}
