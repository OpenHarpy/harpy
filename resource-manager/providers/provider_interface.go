package providers

import (
	"resource-manager/logger"
	obj "resource-manager/objects"
)

// A provider interface is a contract that all providers must implement
// This interface is used to abstract the provider implementation from the resource manager
// This allows the resource manager to be provider agnostic
// Harpy aims to abstract any provider implementation, to that effect, we make it part of our design the following:
// - Providers are responsible for managing the lifecycle of the nodes that they provision.
type ProviderProvisionResponse struct {
	NodeID          string
	NodeGRPCAddress string
}

type ProviderDecommissionResponse struct {
	NodeID string
}

type ProviderProps struct {
	ProviderName          string
	ProviderDescription   string
	ProviderCorePerNode   int
	ProviderMemoryPerNode int
	ProviderWarmpoolSize  int
	ProviderMaxNodes      int
}

type ProviderInterface interface {
	// Begin will initialize the provider
	Begin() error // This is supposed to be a blocking call and it will be called by Init
	// ProvisionNode will provision a node of the specified type
	ProvisionNodes(nodeCount int) ([]*ProviderProvisionResponse, error)
	// DestroyNode will destroy a node of the specified type
	DestroyNode(nodeID string) error
	// We need to also have a call to inform the provider if a node has shutdown willingly
	NodeShutdownCallback(nodeID string) error
	// Cleanup will cleanup the provider
	Cleanup() error // IMPORTANT: This is supposed to be a blocking call and it will be called by the resource manager
	// ProviderTick will be called by the resource manager at regular intervals to allow the provider to do any housekeeping
	ProviderTick() ([]*ProviderDecommissionResponse, []*ProviderProvisionResponse, error)
	// Other functions
	GeneratedProviderDescription() ProviderProps
	// Check if the provider can autoscale
	CanAutoScale() bool
}

func StartProvider(p ProviderInterface) error {
	logger.Info("Init called", "PROVIDER_INTERFACE")
	obj.ResetProvider()

	providerDef := p.GeneratedProviderDescription()
	providerObj := obj.Provider{
		ProviderName:          providerDef.ProviderName,
		ProviderDescription:   providerDef.ProviderDescription,
		ProviderCorePerNode:   providerDef.ProviderCorePerNode,
		ProviderMemoryPerNode: providerDef.ProviderMemoryPerNode,
		ProviderWarmpoolSize:  providerDef.ProviderWarmpoolSize,
		ProviderMaxNodes:      providerDef.ProviderMaxNodes,
	}
	providerObj.Sync()

	err := p.Begin()
	if err != nil {
		return err
	}
	return nil
}
