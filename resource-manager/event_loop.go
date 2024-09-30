package main

import (
	"resource-manager/logger"
	"time"

	"github.com/google/uuid"
)

// The interval at which the process event loop will check for processes
const processPoolingInterval = 1 * time.Second

func SpinupNode(lm *LiveMemory, nodeType string, nodeCount int) []string {
	// This function will spin up a node of the specified type
	// The node will be added to the pool
	// TODO: Add logic to spin up the node
	// For now we will just add an id to the pool
	nodeIDs := []string{}
	for i := 0; i < nodeCount; i++ {
		nodeId := uuid.New().String()
		nodeIDs = append(nodeIDs, nodeId)
		liveNode := &LiveNode{
			NodeID:          nodeId,
			NodeType:        nodeType,
			NodeGRPCAddress: "",
			NodeStatus:      LiveNodeStatusEnum_NODE_STATUS_UNKNOWN,
		}
		lm.NodePool[nodeId] = liveNode
		lm.NodeTypePool[nodeType] = append(lm.NodeTypePool[nodeType], &nodeId)
	}
	return nodeIDs
}

func EvalNodeAllocation(lm *LiveMemory) {
	// This function will loop through all the nodes and check if they are idle
	// If they are idle then it will check if the idle time has passed
	// If the idle time has passed then it will remove the node from the pool
	nodeRequests := lm.ResourceAssignments

	for _, nodeRequest := range nodeRequests {
		// Check if the there are requests that are not yet fulfilled
		if nodeRequest.ServingStatus == ResourceRequestStatusEnum_RESOURCE_REQUESTED {
			// Check if there are nodes of the requested type available in the pool
			// If there are nodes available then assign the nodes to the request
			requestedType := nodeRequest.NodeType
			requestedCount := nodeRequest.NodeCount
			nodesInPool := lm.NodeTypePool[requestedType]
			availableNodes := []string{}
			for _, nodeID := range nodesInPool {
				if lm.NodeServingRequests[*nodeID] == nil {
					availableNodes = append(availableNodes, *nodeID)
				}
			}

			// Check: If there are not enough nodes available then spin up the required number of nodes
			if len(availableNodes) < int(requestedCount) {
				logger.Info("Not enough nodes available", "EVAL_NODE_ALLOCATION")
				nodeIDs := SpinupNode(lm, requestedType, int(requestedCount)-len(availableNodes))
				// Tag these nodes as serving this request
				for _, nodeID := range nodeIDs {
					lm.NodeServingRequests[nodeID] = &nodeRequest.RequestID
				}
			} else {
				// Assign the nodes to the request
				for i := 0; i < int(requestedCount); i++ {
					lm.NodeServingRequests[availableNodes[i]] = &nodeRequest.RequestID
					nodeRequest.LiveNodes = append(nodeRequest.LiveNodes, availableNodes[i])
				}
			}
			nodeRequest.ServingStatus = ResourceRequestStatusEnum_RESOURCE_ALLOCATING
		}
	}
}

func EvalNodesReady(lm *LiveMemory) {
	// This function will get any request that is in the RESOURCE_ALLOCATING state and check if the nodes are ready
	// If all the nodes are ready then we will transition the request to RESOURCE_ALLOCATED
	for _, nodeRequest := range lm.ResourceAssignments {
		if nodeRequest.ServingStatus == ResourceRequestStatusEnum_RESOURCE_ALLOCATING {
			// Check if all the nodes are ready
			ready := true
			for _, nodeID := range nodeRequest.LiveNodes {
				node := lm.NodePool[nodeID]
				if node.NodeStatus != LiveNodeStatusEnum_NODE_STATUS_READY {
					ready = false
					break
				}
			}
			if ready {
				nodeRequest.ServingStatus = ResourceRequestStatusEnum_RESOURCE_ALLOCATED
			}
		}
	}
}

func EvalResourceReleaseRequest(lm *LiveMemory) {
	// This function will loop through all the requests and check if they are in the RESOURCE_RELEASE_REQUESTED state
	// If they are in the RESOURCE_RELEASE_REQUESTED state then it will release the nodes
	for _, nodeRequest := range lm.ResourceAssignments {
		if nodeRequest.ServingStatus == ResourceRequestStatusEnum_RESOURCE_RELEASE_REQUESTED {
			// Release the nodes
			for _, nodeID := range nodeRequest.LiveNodes {
				// Remove the node mapping from the serving requests
				lm.NodeServingRequests[nodeID] = nil
				// We now remove the node from the ServingRequest
				nodeRequest.LiveNodes = []string{}
			}
			// Transition the request to RESOURCE_RELEASED
			nodeRequest.ServingStatus = ResourceRequestStatusEnum_RESOURCE_RELEASED
		}
	}
}

func ProcessEventLoop(existEventLoop chan bool, waitEventLoopExitChan chan bool, lm *LiveMemory) {
	logger.Info("Starting process event loop", "PROCESS_EVENT_LOOP")
	// This function will loop through all the processes and check if they are done
	// If they are done then it will set the process status to done
	for {
		select {
		case <-existEventLoop:
			// Cleanup code here
			logger.Info("Exiting event loop", "PROCESS_EVENT_LOOP")
			waitEventLoopExitChan <- true
			return
		default:
			// Current running processes
			EvalNodeAllocation(lm)
			EvalNodesReady(lm)
			EvalResourceReleaseRequest(lm)
			// TODO: We need to add logic here to timeout the nodes and requests that are not being used for a long time
			// TODO: We also need to implement logic to handle panics and nodes that are not able to come up
			time.Sleep(processPoolingInterval)
		}
	}
}
