// Package task implements any task related operations
//
// This file contains the implementation of the node scheduler
//

package task

import (
	"client-engine/logger"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

/* NodeScheduler */
const (
	NODE_SCHEDULER_HEARTBEAT_INTERVAL = 30 * time.Second
	NODE_SCHEDULER_POOLING_INTERVAL   = 500 * time.Millisecond
)

// ENUM for the Scheduler algorithm
type SchedulerAlgorithm int

const (
	SCHEDULER_ALGORITHM_ROUND_ROBIN SchedulerAlgorithm = 0
)

type Scheduler interface {
	GetNextNodeID() int
}

type RoundRobinScheduler struct {
	CurrentIndex int
	MaxIndex     int
}

func (r *RoundRobinScheduler) GetNextNodeID() int {
	r.CurrentIndex++
	if r.CurrentIndex >= r.MaxIndex {
		r.CurrentIndex = 0
	}
	return r.CurrentIndex
}

// NodeScheduler

type NodeScheduler struct {
	NodesList                       []*Node
	ResourceManager                 *ResourceManagerClient
	ResourceRequestResponse         *NodeRequestResponse
	NodeSchedulerHeartbeatExit      chan bool
	NodeSchedulerHeartbeatWaitGroup *sync.WaitGroup
	BlockTracker                    *BlockTracker
	Scheduler                       Scheduler
	SchedulerAlgorithm              SchedulerAlgorithm
}

func NewNodeScheduler(CallbackURI string, ResourceManagerURI string, SessionID string, NodeCount int) (*NodeScheduler, error) {
	// Start a new BlockTracker instance
	blockTracker := NewBlockTracker()
	// We will need to get the nodes from the resource manager
	resourceManager := NewNodeResourceManagerClient(ResourceManagerURI)
	resourceManager.connect()
	requestResponse, err := resourceManager.RequestNode(NodeCount)
	logger.Debug("Requesting node", "NODE-TRACKER", logrus.Fields{"RequestID": requestResponse.RequestID, "nodeCount": requestResponse.RequestStatus})

	if err != nil {
		logger.Error("Failed to request node", "NODE-TRACKER", err)
		return nil, err
	}

	// Now we pool for the nodes to be ready
	for {
		// We will need to get the nodes from the resource manager
		requestResponse, err = resourceManager.GetNodeRequestStatus(requestResponse)
		if err != nil {
			logger.Error("Failed to get request status", "NODE-TRACKER", err)
			return nil, err
		}

		if requestResponse.RequestStatus == REQUEST_ALLOCATED {
			logger.Debug("Nodes allocated", "NODE-TRACKER")
			break
		} else if requestResponse.RequestStatus == REQUEST_ERROR {
			resourceManager.ReleaseNodes(requestResponse)
			resourceManager.disconnect()
			return nil, errors.New("failed to get nodes")
		}
		time.Sleep(NODE_SCHEDULER_POOLING_INTERVAL)
	}

	// We now add the nodes to the tracker
	NodeScheduler := &NodeScheduler{
		NodesList:               make([]*Node, 0),
		ResourceManager:         resourceManager,
		ResourceRequestResponse: requestResponse,
		BlockTracker:            blockTracker,
		Scheduler:               &RoundRobinScheduler{CurrentIndex: 0, MaxIndex: len(requestResponse.Nodes)}, // Default to round-robin
		SchedulerAlgorithm:      SCHEDULER_ALGORITHM_ROUND_ROBIN,                                             // Default to round-robin
	}

	for _, liveNode := range requestResponse.Nodes {
		// Add the node to the tracker (this will connect the node, register the callback and initialize the isolated environment)
		err = NodeScheduler.AddNode(liveNode.NodeGRPCAddress, CallbackURI, SessionID)
		if err != nil {
			logger.Error("Failed to add node", "NODE-TRACKER", err, logrus.Fields{"nodeID": liveNode.NodeID, "nodeGRPCAddress": liveNode.NodeGRPCAddress})
			resourceManager.ReleaseNodes(requestResponse)
			resourceManager.disconnect()
			return nil, err
		}
	}

	// Everything is ready, now we can start the heartbeat
	NodeScheduler.NodeSchedulerHeartbeatExit = make(chan bool)
	NodeScheduler.NodeSchedulerHeartbeatWaitGroup = &sync.WaitGroup{}
	NodeScheduler.NodeSchedulerHeartbeatWaitGroup.Add(1)
	go NodeScheduler.HeartbeatRoutine()
	return NodeScheduler, nil
}

func (n *NodeScheduler) HeartbeatRoutine() {
	// Send the first heartbeat
	n.ResourceManager.SendHeartbeat(n.ResourceRequestResponse.RequestID)
	// Send a heartbeat every NODE_SCHEDULER_HEARTBEAT_INTERVAL
	for {
		select {
		case <-n.NodeSchedulerHeartbeatExit:
			defer n.NodeSchedulerHeartbeatWaitGroup.Done()
			return
		case <-time.After(NODE_SCHEDULER_HEARTBEAT_INTERVAL):
			n.ResourceManager.SendHeartbeat(n.ResourceRequestResponse.RequestID)
		}
	}
}

func (n *NodeScheduler) AddNode(nodeURI string, callbackURI string, SessionID string) error {
	node := &Node{
		nodeID:       uuid.New().String(),
		NodeURI:      nodeURI,
		CallbackURI:  callbackURI,
		SessionID:    SessionID,
		BlockTracker: n.BlockTracker,
	}
	err := node.connect()
	if err != nil {
		return err
	}
	n.NodesList = append(n.NodesList, node) // Add the node to the list
	return nil
}

func (n *NodeScheduler) GetNextNode() *Node {
	if len(n.NodesList) == 0 {
		return nil
	}
	// Get the next node
	nodeID := n.Scheduler.GetNextNodeID()
	node := n.NodesList[nodeID]
	return node
}

func (n *NodeScheduler) GetNode(nodeID string) *Node {
	for _, node := range n.NodesList {
		if node.nodeID == nodeID {
			return node
		}
	}
	return nil
}

func (n *NodeScheduler) Close() {
	for _, node := range n.NodesList {
		node.disconnect()
	}
	n.ResourceManager.ReleaseNodes(n.ResourceRequestResponse)
	n.ResourceManager.disconnect()
	n.NodeSchedulerHeartbeatExit <- true
	n.NodeSchedulerHeartbeatWaitGroup.Wait() // Wait for the heartbeat to finish
}

func (n *NodeScheduler) FlushBlocks(GroupID string, filterOutBlockTypes []string) {
	node := n.GetNextNode() // Get any node to flush the blocks (we plan to use a distributed file system in the future)
	blocksToFlush := n.BlockTracker.GetBlockGroup(GroupID, filterOutBlockTypes)
	if blocksToFlush == nil {
		logger.Warn("No blocks to flush", "NODE-TRACKER", logrus.Fields{"GroupID": GroupID})
		return
	}
	logger.Info("Flushing block", "NODE-TRACKER", logrus.Fields{"BlocksToFlush": len(blocksToFlush)})
	for _, block := range blocksToFlush {
		err := node.FlushBlock(block.BlockID)
		if err != nil {
			logger.Warn("Failed to flush block, this can be normal depending on the taskset setup", "NODE-TRACKER", logrus.Fields{"blockID": block.BlockID})
		}
	}
}

func (n *NodeScheduler) SwapSchedulingAlgorithm(algorithm SchedulerAlgorithm) error {
	n.SchedulerAlgorithm = algorithm
	switch algorithm {
	case SCHEDULER_ALGORITHM_ROUND_ROBIN:
		n.Scheduler = &RoundRobinScheduler{CurrentIndex: 0, MaxIndex: len(n.NodesList)}
	default:
		return errors.New("unknown algorithm")
	}
	return nil
}
