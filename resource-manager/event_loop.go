package main

import (
	"resource-manager/logger"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// The interval at which the process event loop will check for processes
const processPoolingInterval = 1 * time.Second
const resourceRequestReleasedTimeout = 60 * time.Second

func SpinupNode(nodeType string, nodeCount int) []string {
	// This function will spin up a node of the specified type
	// The node will be added to the pool
	// TODO: Add logic to spin up the node
	return []string{}
}

func EvalNodeAllocation() {
	// This function will loop through all the nodes and check if they are idle
	// If they are idle then it will check if the idle time has passed
	// If the idle time has passed then it will remove the node from the pool
	//nodeRequests := lm.ResourceAssignments
	nodeRequests := GetResourceAssignmentsByStatus(ResourceRequestStatusEnum_RESOURCE_REQUESTED)

	for _, nodeRequest := range nodeRequests {
		// Check if there are nodes of the requested type available in the pool
		// If there are nodes available then assign the nodes to the request
		requestedType := nodeRequest.NodeType
		requestedCount := nodeRequest.NodeCount
		availableNodes := GetLiveNodesNotServingRequest(requestedType)

		// Check: If there are not enough nodes available then spin up the required number of nodes
		if len(availableNodes) < int(requestedCount) {
			logger.Info("Not enough nodes available", "EVAL_NODE_ALLOCATION")
			//nodeIDs := SpinupNode(lm, requestedType, int(requestedCount)-len(availableNodes))
			// Tag these nodes as serving this request
			//for _, nodeID := range nodeIDs {
			//	lm.NodeServingRequests[nodeID] = &nodeRequest.RequestID
			//}
			// TODO: Here is where we would call the provider to spin up the nodes
		} else {
			// Assign the nodes to the request
			for i := 0; i < int(requestedCount); i++ {
				node := availableNodes[i]
				node.ServingRequestID = nodeRequest.RequestID
				node.IsServingRequest = true
				node.Sync()
			}
		}
		nodeRequest.ServingStatus = ResourceRequestStatusEnum_RESOURCE_ALLOCATING
		nodeRequest.Sync()
	}
}

func EvalNodesReady() {
	// This function will get any request that is in the RESOURCE_ALLOCATING state and check if the nodes are ready
	// If all the nodes are ready then we will transition the request to RESOURCE_ALLOCATED
	nodeRequests := GetResourceAssignmentsByStatus(ResourceRequestStatusEnum_RESOURCE_ALLOCATING)
	for _, nodeRequest := range nodeRequests {
		// Check if all the nodes are ready
		ready := true
		nodesServingThisRequest := GetLiveNodesServingRequest(nodeRequest.RequestID)
		for _, node := range nodesServingThisRequest {
			if node.NodeStatus != LiveNodeStatusEnum_NODE_STATUS_READY {
				ready = false
				break
			}
		}
		if ready {
			nodeRequest.ServingStatus = ResourceRequestStatusEnum_RESOURCE_ALLOCATED
			nodeRequest.Sync()
		}
	}
}

func EvalResourceReleaseRequest() {
	// This function will loop through all the requests and check if they are in the RESOURCE_RELEASE_REQUESTED state
	// If they are in the RESOURCE_RELEASE_REQUESTED state then it will release the nodes
	nodeRequests := GetResourceAssignmentsByStatus(ResourceRequestStatusEnum_RESOURCE_RELEASE_REQUESTED)
	for _, nodeRequest := range nodeRequests {
		if nodeRequest.ServingStatus == ResourceRequestStatusEnum_RESOURCE_RELEASE_REQUESTED {
			// Release the nodes
			nodesServing := GetLiveNodesServingRequest(nodeRequest.RequestID)
			for _, node := range nodesServing {
				// Remove the node mapping from the serving requests
				node.ServingRequestID = ""
				node.IsServingRequest = false
				node.Sync()
			}
			// We now remove the node from the ServingRequest
			// Transition the request to RESOURCE_RELEASED
			nodeRequest.ServingStatus = ResourceRequestStatusEnum_RESOURCE_RELEASED
			nodeRequest.Sync()
		}
	}
}

func EvalResourceRequestReleasedTimeout() {
	// This function will loop through all the requests and check if they are in the RESOURCE_RELEASED state
	// If their heartbeat crosses a certain threshold then we will remove the request from the database
	nodeRequests := GetResourceAssignmentsByStatus(ResourceRequestStatusEnum_RESOURCE_RELEASED)
	for _, nodeRequest := range nodeRequests {
		// Check if the heartbeat has crossed a certain threshold
		// If it has then we will remove the request from the database
		timeSince := time.Since(time.Unix(nodeRequest.LastHeartbeatReceived, 0))
		if timeSince > resourceRequestReleasedTimeout {
			logger.Info("Request has been released for too long", "EVAL_RESOURCE_REQUEST_RELEASED_TIMEOUT", logrus.Fields{"request_id": nodeRequest.RequestID, "time_since": timeSince})
			// Remove the request from the database
			nodeRequest.Delete()
		}
	}
}

func DoProcessLoop(exitEventLoop chan bool, wg *sync.WaitGroup) {
	logger.Info("Starting process event loop", "PROCESS_EVENT_LOOP")
	// This function will loop through all the processes and check if they are done
	// If they are done then it will set the process status to done
	for {
		select {
		case <-exitEventLoop:
			// Cleanup code here
			logger.Info("Exiting event loop", "PROCESS_EVENT_LOOP")
			defer wg.Done()
			return
		default:
			// Current running processes
			EvalNodeAllocation()
			EvalNodesReady()
			EvalResourceReleaseRequest()
			// TODO: We need to add logic here to timeout the nodes and requests that are not being used for a long time
			// TODO: We also need to implement logic to handle panics and nodes that are not able to come up
			EvalResourceRequestReleasedTimeout()
			time.Sleep(processPoolingInterval)
		}
	}
}

func ProcessEventLoop(existEventLoop chan bool, wg *sync.WaitGroup) error {
	go DoProcessLoop(existEventLoop, wg)
	return nil
}
