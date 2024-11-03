package providers

import (
	"resource-manager/logger"
	obj "resource-manager/objects"
)

// A provider interface is a contract that all providers must implement
// This interface is used to abstract the provider implementation from the resource manager
// This allows the resource manager to be provider agnostic
// Harpy aims to abstract any provider implementation, to that effect, we make it part of our design the following:
// - Providers support nodeType as inputs. This is a string that represents the type of node that is being requested.
// - By design the naming of the nodeType MUST NOT be provider specific.
// - In this case providers can return a set of node types that they support but those need to follow the naming convention
// - For example: AWS has a nodeType called "t2.micro" but the provider implementation should return "small-1cpu-1gb" instead of "t2.micro"
// --- Improvements of cost and performance can be made by the provider implementation but the naming convention should be kept provider agnostic
// --- This is to ensure that if a provider wants, to use, for instance a "t2.medium" to serve two "small-1cpu-1gb", effectively optimizing costs, the implementation can do that as it sees fit
type ProviderProvisionResponse struct {
	NodeType        string
	NodeID          string
	NodeGRPCAddress string
}

type ProviderProps struct {
	ProviderName        string
	ProviderDescription string
}

type ProviderInterface interface {
	// Begin will initialize the provider
	Begin() error // This is supposed to be a blocking call and it will be called by Init
	// ProvisionNode will provision a node of the specified type
	ProvisionNodes(nodeType string, nodeCount int) ([]*ProviderProvisionResponse, error)
	// DestroyNode will destroy a node of the specified type
	DestroyNode(nodeID string) error
	// We need to also have a call to inform the provider if a node has shutdown willingly
	NodeShutdownCallback(nodeID string) error
	// GetNodeTypes will return a list of node types that the provider supports
	GetNodeTypes() ([]string, error)
	// Cleanup will cleanup the provider
	Cleanup() error // IMPORTANT: This is supposed to be a blocking call
	// Other functions
	GeneratedProviderDescription() ProviderProps
	// Check if the provider can autoscale the node type
	CanAutoScale(nodeType string) bool
}

func StartProvider(p ProviderInterface) error {
	logger.Info("Init called", "PROVIDER_INTERFACE")
	providers := obj.GetProviders()
	for _, provider := range providers {
		provider.Delete()
	}

	providerDef := p.GeneratedProviderDescription()
	providerObj := obj.Provider{
		ProviderName:        providerDef.ProviderName,
		ProviderDescription: providerDef.ProviderDescription,
	}
	providerObj.Sync()

	err := p.Begin()
	if err != nil {
		return err
	}
	return nil
}
