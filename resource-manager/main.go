package main

import (
	"os"
	"os/signal"
	"resource-manager/config"
	"resource-manager/logger"
	"syscall"
)

// Begin LiveMemory
type LiveMemory struct {
	// The Live Memory is a struct that holds all the live data of the server
	// It is used to store all the live data of the server

	NodeCatalogs        map[string]*NodeCatalog        // This is a map between nodeType -> NodeCatalog
	NodePool            map[string]*LiveNode           // This is a map between nodeID -> LiveNode
	NodeTypePool        map[string][]*string           // This is a map between nodeType -> nodeID
	NodeServingRequests map[string]*string             // This is a map between nodeID -> requestID
	ResourceAssignments map[string]*ResourceAssignment // This is a map between resourceRequest -> ResourceAssignment
}

// For debugging purposes we will have a function to add some local running nodes
//
//	This will be changed to PROVIDERS in the future
//	Providers will be the implementation of the node instancing and destruction process
//	Examples of providers are AWS, GCP, Azure, local, etc.
func AddLocalNodes(lm *LiveMemory) {
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
	lm.NodeCatalogs["small-4cpu-8gb"] = NodeCatalog
	// This will make it so that nodes of this type can be requested
	// This is a local node
	liveNode := &LiveNode{
		NodeID:          "local-1",
		NodeType:        "small-4cpu-8gb",
		NodeGRPCAddress: "",
		NodeStatus:      LiveNodeStatusEnum_NODE_STATUS_UNKNOWN,
	}
	lm.NodePool["local-1"] = liveNode // This ID will be used to reference this node
	lm.NodeTypePool["small-4cpu-8gb"] = []*string{&liveNode.NodeID}
	// The node will assign itself to the resource manager
	// Upon spinning up the node should "know" its own ID and it should "know" the address of the resource manager
	// This is done by passing the nodeID and the resource manager's address as an argument to the remote-runner's main process
	// As this is just a debug function we will make sure that this is done correctly
	// But in the future this will be done by the provider itself
}

func main() {
	logger.SetupLogging()
	logger.Info("Starting Resource Manager", "MAIN")

	// Make the live memory struct
	lm := &LiveMemory{
		NodeCatalogs:        make(map[string]*NodeCatalog),
		NodePool:            make(map[string]*LiveNode),
		ResourceAssignments: make(map[string]*ResourceAssignment),
		NodeTypePool:        make(map[string][]*string),
		NodeServingRequests: make(map[string]*string),
	}

	// Add some local nodes
	AddLocalNodes(lm)

	// Start the gRPC server
	exitMainServer := make(chan bool)
	waitServerExitChan := make(chan bool)
	port := config.GetConfigs().GetConfigsWithDefault("port", "50050")
	go NewResourceAllocServer(exitMainServer, waitServerExitChan, lm, port)

	// Start the event loop
	exitEventLoop := make(chan bool)
	waitEventLoopExitChan := make(chan bool)
	go ProcessEventLoop(exitEventLoop, waitEventLoopExitChan, lm)

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
