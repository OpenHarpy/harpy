// This file implements the local provider
// The local provider is used for debugging/testing purposes
// It allows the resource manager to create nodes locally
package providers

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"resource-manager/logger"
	obj "resource-manager/objects"
	"resource-manager/providers"
	"strings"

	"github.com/sirupsen/logrus"
)

// LocalProvider is a provider that is used for debugging/testing purposes
// It allows the resource manager to create nodes locally on the same machine as the resource manager
type LocalProvider struct {
	NodesProcessTracker map[string]*exec.Cmd
	CommandToExecute    string
}

func (l *LocalProvider) Begin() error {
	// This function will initialize the provider
	nodes := obj.GetLiveNodes()
	for _, node := range nodes {
		node.Delete()
	}
	return nil
}

func (l *LocalProvider) ProvisionNodes(nodeCount int) ([]*providers.ProviderProvisionResponse, error) {
	logger.Info("Provisioning node", "LOCAL_PROVIDER", logrus.Fields{"nodeCount": nodeCount})
	// This function will provision a node of the specified type
	// The node will be added to the pool
	if nodeCount > 1 {
		return nil, errors.New("local provider can only provision one node at a time")
	}
	//if len(l.NodesProcessTracker) > 0 {
	//	return nil, errors.New("local provider can only provision one node at a time")
	//}
	logger.Info("Starting local node process", "LOCAL_PROVIDER")
	// Start the local process
	port := fmt.Sprintf("%d", 50053+len(l.NodesProcessTracker))
	nodeID := fmt.Sprintf("local-%d", len(l.NodesProcessTracker))

	commandWithPort := strings.Replace(l.CommandToExecute, "{{nodePort}}", port, -1)
	commandWithID := strings.Replace(commandWithPort, "{{nodeID}}", nodeID, -1)
	cmd := exec.Command("sh", "-c", commandWithID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	l.NodesProcessTracker[nodeID] = cmd
	return []*providers.ProviderProvisionResponse{
		{
			NodeID:          nodeID,
			NodeGRPCAddress: fmt.Sprintf("localhost:%s", port),
		},
	}, nil
}

func (l *LocalProvider) DestroyNode(nodeID string) error {
	// This function will destroy a node of the specified type
	// The node will be removed from the pool
	logger.Info("Destroying node", "LOCAL_PROVIDER", logrus.Fields{"nodeID": nodeID})
	// Check if the node is mapped in the pool
	if _, ok := l.NodesProcessTracker[nodeID]; !ok {
		return errors.New("node not found in the pool")
	}
	// Stop the process
	l.NodesProcessTracker[nodeID].Process.Signal(os.Interrupt)
	l.NodesProcessTracker[nodeID].Wait()
	delete(l.NodesProcessTracker, nodeID)
	return nil
}

func (l *LocalProvider) NodeShutdownCallback(nodeID string) error {
	// This function will be called by the resource manager when a node has shutdown willingly
	// The provider should remove the node from the pool
	logger.Info("Node shutdown callback", "LOCAL_PROVIDER", logrus.Fields{"nodeID": nodeID})
	// Check if the node is mapped in the pool
	if _, ok := l.NodesProcessTracker[nodeID]; !ok {
		return errors.New("node not found in the pool")
	}
	// Stop the process
	l.NodesProcessTracker[nodeID].Process.Signal(os.Interrupt)
	l.NodesProcessTracker[nodeID].Wait()
	delete(l.NodesProcessTracker, nodeID)
	return nil
}

func (l *LocalProvider) ProviderTick() ([]*providers.ProviderDecommissionResponse, []*providers.ProviderProvisionResponse, error) {
	// This function will be called by the resource manager at regular intervals to allow the provider to do any housekeeping
	if len(l.NodesProcessTracker) == 0 {
		// Start the local process
		provisionResults, err := l.ProvisionNodes(1)
		if err != nil {
			return nil, nil, err
		}
		return nil, provisionResults, nil
	}

	return nil, nil, nil
}

func (l *LocalProvider) Cleanup() error {
	logger.Info("Cleaning up local provider", "LOCAL_PROVIDER")
	// This function will cleanup the provider
	for _, cmd := range l.NodesProcessTracker {
		cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}
	return nil
}

func (l *LocalProvider) GeneratedProviderDescription() providers.ProviderProps {
	return providers.ProviderProps{
		ProviderName:          "local",
		ProviderDescription:   "Local provider for debugging/testing purposes",
		ProviderCorePerNode:   4,
		ProviderMemoryPerNode: 2048,
		ProviderWarmpoolSize:  1,
	}
}

func (l *LocalProvider) CanAutoScale() bool {
	// This function will check if the provider can autoscale
	return false
}

func NewLocalProvider(CommandToExecute string) *LocalProvider {
	return &LocalProvider{
		CommandToExecute:    CommandToExecute,
		NodesProcessTracker: make(map[string]*exec.Cmd),
	}
}
