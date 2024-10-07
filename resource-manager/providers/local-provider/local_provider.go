// This file implements the local provider
// The local provider is used for debugging/testing purposes
// It allows the resource manager to create nodes locally
package providers

import (
	"errors"
	"os"
	"os/exec"
	"resource-manager/logger"
	obj "resource-manager/objects"
	"resource-manager/providers"
	"sync"

	"github.com/sirupsen/logrus"
)

// LocalProvider is a provider that is used for debugging/testing purposes
// It allows the resource manager to create nodes locally on the same machine as the resource manager

const commandToStartLocalNode = "cd ../remote-runner && go run . local-1 small-4cpu-8gb localhost:50050"

func CreateAndManageLocalProcess(exitChan chan bool, wg *sync.WaitGroup) {
	logger.Info("Starting local node process", "CREATE_MANAGE_LOCAL_PROCESS")
	// This function will be called as a goroutine and will manage the local process
	cmd := exec.Command("sh", "-c", commandToStartLocalNode)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		logger.Error("Failed to start local node", "CREATE_MANAGE_LOCAL_PROCESS", err)
		// TODO: ADD ERROR HANDLING (PROVIDERS SHOULD HAVE AN "ERROR QUEUE")
		defer wg.Done()
		return
	}
	// We need to check if the exitChan has been closed if so we need to send ctrl+c (SIGINT) to the process
	go func() {
		<-exitChan
		cmd.Process.Signal(os.Interrupt)
	}()
	go func() {
		cmd.Wait()
		// Print status of the process
		logger.Info("Local node process exited", "CREATE_MANAGE_LOCAL_PROCESS")
	}()
	// Create a goroutine to continuously read the stdout and stderr of the process
	err = cmd.Wait()
	if err != nil {
		logger.Error("Local node process exited with error", "CREATE_MANAGE_LOCAL_PROCESS", err)
	}
	defer wg.Done()
}

func StopLocalProcess(exitChan chan bool, wg *sync.WaitGroup, lp *LocalProvider) {
	// This function will stop the local process
	if lp.NodeAlreadyProvisioned {
		exitChan <- true
		wg.Wait()
		lp.NodeAlreadyProvisioned = false
	}
}

type LocalProvider struct {
	NodeAlreadyProvisioned bool
	ExitChan               chan bool
	Wg                     *sync.WaitGroup
}

func (l *LocalProvider) Begin() error {
	// This function will initialize the provider
	nodes := obj.GetLiveNodes()
	for _, node := range nodes {
		node.Delete()
	}
	return nil
}

func (l *LocalProvider) ProvisionNodes(nodeType string, nodeCount int) ([]*providers.ProviderProvisionResponse, error) {
	logger.Info("Provisioning node", "PROVISION_NODES", logrus.Fields{"nodeType": nodeType, "nodeCount": nodeCount})
	// This function will provision a node of the specified type
	// The node will be added to the pool
	if nodeType != "small-4cpu-8gb" {
		return nil, errors.New("node type not supported by local provider")
	}
	if nodeCount > 1 {
		return nil, errors.New("local provider can only provision one node at a time")
	}
	if l.NodeAlreadyProvisioned {
		return nil, errors.New("local provider cannot provision more than one node")
	}
	l.NodeAlreadyProvisioned = true
	// We will start the local process
	exitChan := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(1)
	l.ExitChan = exitChan
	l.Wg = &wg
	logger.Info("Starting local node process", "PROVISION_NODES")
	go CreateAndManageLocalProcess(exitChan, &wg)

	return []*providers.ProviderProvisionResponse{
		{
			NodeType:        "small-4cpu-8gb",
			NodeID:          "local-1",
			NodeGRPCAddress: "localhost:50051",
		},
	}, nil
}

func (l *LocalProvider) DestroyNode(nodeID string) error {
	// This function will destroy a node of the specified type
	// The node will be removed from the pool
	if nodeID != "local-1" {
		return errors.New("node not found")
	} else {
		go StopLocalProcess(l.ExitChan, l.Wg, l)
		return nil
	}
}

func (l *LocalProvider) NodeShutdownCallback(nodeID string) error {
	// This function will be called by the resource manager when a node has shutdown willingly
	// The provider should remove the node from the pool
	if nodeID != "local-1" {
		return errors.New("node not found")
	} else {
		go StopLocalProcess(l.ExitChan, l.Wg, l)
		return nil
	}
}

func (l LocalProvider) GetNodeTypes() ([]string, error) {
	// This function will return a list of node types that the provider supports
	return []string{"small-4cpu-8gb"}, nil
}

func (l *LocalProvider) Cleanup() error {
	// This function will cleanup the provider
	StopLocalProcess(l.ExitChan, l.Wg, l)
	return nil
}

func (l *LocalProvider) GeneratedProviderDescription() providers.ProviderProps {
	return providers.ProviderProps{
		ProviderName:        "local",
		ProviderDescription: "Local provider for debugging/testing purposes",
	}
}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{}
}
