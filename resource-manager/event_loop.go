package main

import (
	"resource-manager/logger"
	obj "resource-manager/objects"
	"resource-manager/providers"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// The interval at which the process event loop will check for processes
var processPoolingInterval = 1 * time.Second
var resourceRequestReleasedTimeout = 60 * time.Second
var requestIdleTimeout = 120 * time.Second

func SpinupNode(runningProvider providers.ProviderInterface, count int) {
	// This function will spin up a node of the specified type
	pr, err := runningProvider.ProvisionNodes(count)
	if err != nil {
		logger.Error("Failed to spin up node", "SPINUP_NODE", err)
	}
	// Add the nodes
	for _, node := range pr {
		liveNode := &obj.LiveNode{
			NodeID:           node.NodeID,
			NodeStatus:       obj.LiveNodeStatusEnum_NODE_STATUS_BOOTING,
			NodeGRPCAddress:  node.NodeGRPCAddress,
			IsServingRequest: false,
			ServingRequestID: "",
			NodeCreatedAt:    time.Now().Unix(),
		}
		liveNode.Sync()
	}
}

func EvalProviderTick(runningProvider providers.ProviderInterface) {
	// Provider ticks are used to push the provider to do some work, like decommissioning nodes, starting nodes, etc
	// The provider tick will return the nodes that have been decommissioned and the nodes that have been started
	// We will sync the nodes in the database with the nodes that have been decommissioned and started
	nodesDecommissioned, nodesStarted, err := runningProvider.ProviderTick()
	if err != nil {
		logger.Error("Provider tick failed", "EVAL_TICK", err)
	}
	// Sync the nodes
	for _, node := range nodesDecommissioned {
		logger.Info("Decommissioning node", "EVAL_TICK", logrus.Fields{"node_id": node.NodeID})
		liveNode, exists := obj.GetLiveNode(node.NodeID)
		if exists {
			liveNode.Delete()
		}
	}
	for _, node := range nodesStarted {
		logger.Info("Starting node", "EVAL_TICK", logrus.Fields{"node_id": node.NodeID})
		liveNode := &obj.LiveNode{
			NodeID:           node.NodeID,
			NodeStatus:       obj.LiveNodeStatusEnum_NODE_STATUS_BOOTING,
			NodeGRPCAddress:  node.NodeGRPCAddress,
			IsServingRequest: false,
			ServingRequestID: "",
			NodeCreatedAt:    time.Now().Unix(),
		}
		liveNode.Sync()
	}
}

func EvalNodeAllocation(runningProvider providers.ProviderInterface) {
	// This function will loop through all the nodes and check if they are idle
	// If they are idle then it will check if the idle time has passed
	// If the idle time has passed then it will remove the node from the pool
	//nodeRequests := lm.ResourceAssignments
	nodeRequests := obj.GetResourceAssignmentsByStatus(obj.ResourceAssignmentStatusEnum_RESOURCE_REQUESTED)

	for _, nodeRequest := range nodeRequests {
		// Check if there are nodes of the requested type available in the pool
		// If there are nodes available then assign the nodes to the request
		requestedCount := nodeRequest.NodeCount
		availableNodes := obj.GetLiveNodesNotServingRequest()

		// Check: If there are not enough nodes available then spin up the required number of nodes
		if len(availableNodes) < int(requestedCount) {
			logger.Info("Not enough nodes available", "EVAL_NODE_ALLOCATION")
			// Autoscalle signal
			if runningProvider.CanAutoScale() {
				SpinupNode(runningProvider, int(requestedCount)-len(availableNodes))
			}
			// Else we do nothing and skip this request for now (we will check again in the next iteration)
			continue
		} else {
			// Assign the nodes to the request
			for i := 0; i < int(requestedCount); i++ {
				node := availableNodes[i]
				node.ServingRequestID = nodeRequest.RequestID
				node.IsServingRequest = true
				node.Sync()
			}
			// We have found the nodes for this request, now we transition the request to RESOURCE_ALLOCATING
			nodeRequest.ServingStatus = obj.ResourceAssignmentStatusEnum_RESOURCE_ALLOCATING
			nodeRequest.Sync()
		}
	}
}

func EvalNodesReady() {
	// This function will get any request that is in the RESOURCE_ALLOCATING state and check if the nodes are ready
	// If all the nodes are ready then we will transition the request to RESOURCE_ALLOCATED
	//  Nodes not ready can mean that the node is still booting up
	nodeRequests := obj.GetResourceAssignmentsByStatus(obj.ResourceAssignmentStatusEnum_RESOURCE_ALLOCATING)
	for _, nodeRequest := range nodeRequests {
		// Check if all the nodes are ready
		nodesServingThisRequest := obj.GetLiveNodesServingRequest(nodeRequest.RequestID)
		// if there are no nodes serving this request then we do an early exit of this nodeRequest
		if len(nodesServingThisRequest) == 0 {
			continue
		}
		ready := true
		for _, node := range nodesServingThisRequest {
			if node.NodeStatus != obj.LiveNodeStatusEnum_NODE_STATUS_READY {
				ready = false
				break
			}
		}
		if ready {
			nodeRequest.ServingStatus = obj.ResourceAssignmentStatusEnum_RESOURCE_ALLOCATED
			nodeRequest.Sync()
		}
	}
}

func EvalResourceReleaseRequest() {
	// This function will loop through all the requests and check if they are in the RESOURCE_RELEASE_REQUESTED state
	// If they are in the RESOURCE_RELEASE_REQUESTED state then it will release the nodes
	nodeRequests := obj.GetResourceAssignmentsByStatus(obj.ResourceAssignmentStatusEnum_RESOURCE_RELEASE_REQUESTED)
	for _, nodeRequest := range nodeRequests {
		if nodeRequest.ServingStatus == obj.ResourceAssignmentStatusEnum_RESOURCE_RELEASE_REQUESTED {
			// Release the nodes
			nodesServing := obj.GetLiveNodesServingRequest(nodeRequest.RequestID)
			for _, node := range nodesServing {
				// Remove the node mapping from the serving requests
				node.ServingRequestID = ""
				node.IsServingRequest = false
				node.Sync()
			}
			// We now remove the node from the ServingRequest
			// Transition the request to RESOURCE_RELEASED
			nodeRequest.ServingStatus = obj.ResourceAssignmentStatusEnum_RESOURCE_RELEASED
			nodeRequest.Sync()
		}
	}
}

func EvalResourceRequestTimeout() {
	// This function will loop through all the requests and check if they are in the RESOURCE_RELEASED state
	// If their heartbeat crosses a certain threshold then we will remove the request from the database
	threshold := time.Now().Unix() - int64(resourceRequestReleasedTimeout.Seconds())
	nodeRequests := obj.GetResourceAssignmentsWithHeartbeatThreshold(threshold)
	for _, nodeRequest := range nodeRequests {
		logger.Info("Request has been released for too long", "EVAL_RESOURCE_REQUEST_RELEASED_TIMEOUT", logrus.Fields{"request_id": nodeRequest.RequestID, "threshold": threshold})
		nodeRequest.Delete()
	}
}

func EvalRequestIdleTimeout() {
	// This function will loop through all the requests and check if they are in the RESOURCE_ALLOCATED state
	// If their heartbeat crosses a certain threshold then we will remove the request from the database
	nodeRequests := obj.GetResourceAssignmentsByStatus(obj.ResourceAssignmentStatusEnum_RESOURCE_ALLOCATED)
	for _, nodeRequest := range nodeRequests {
		// Check if the heartbeat has crossed a certain threshold
		// If it has then we will remove the request from the database
		timeSince := time.Since(time.Unix(nodeRequest.LastHeartbeatReceived, 0))
		if nodeRequest.LastHeartbeatReceived == 0 {
			// If the heartbeat was never received we can consider the creation time to compute the time since
			timeSince = time.Since(time.Unix(nodeRequest.RequestCreatedAt, 0))
		}
		if timeSince > requestIdleTimeout {
			logger.Info("Request has been idle for too long", "EVAL_REQUEST_IDLE_TIMEOUT", logrus.Fields{"request_id": nodeRequest.RequestID, "time_since": timeSince})
			// Transition the request to RESOURCE_RELEASE_REQUESTED
			nodeRequest.ServingStatus = obj.ResourceAssignmentStatusEnum_RESOURCE_RELEASE_REQUESTED
			nodeRequest.Sync()
		}
	}
}

func PopulateFromConfigs(key string, DefaultValue time.Duration, multiplier time.Duration) time.Duration {
	configProcessPoolingInterval, ok := obj.GetConfig(key)
	if ok {
		configValue := configProcessPoolingInterval.ConfigValue
		castDuration, err := strconv.ParseInt(configValue, 10, 64)
		if err != nil {
			logger.Warn("Failed to parse "+key+". Reverting to default value", "PROCESS_EVENT_LOOP")
		} else {
			logger.Info("Setting "+key, "PROCESS_EVENT_LOOP")
			return time.Duration(castDuration) * multiplier
		}
	}
	return DefaultValue
}

func EvalNodeShutdownCallbackProvider(runningProvider providers.ProviderInterface) {
	// This function will loop through all the nodes and check if they are in the NODE_STATUS_SHUTTING_DOWN state
	// If they are in the NODE_STATUS_SHUTTING_DOWN state then it will check if the node is done
	// If the node is done then it will remove the node from the pool
	liveNodes := obj.GetLiveNodesByStatus(obj.LiveNodeStatusEnum_NODE_STATUS_SHUTDOWN)
	for _, node := range liveNodes {
		// Check if the node is done
		nodeShutdownError := runningProvider.NodeShutdownCallback(node.NodeID)
		if nodeShutdownError == nil {
			// Remove the node from the pool
			node.Delete()
		}
	}
}

func DoProcessLoop(runningProvider providers.ProviderInterface, exitEventLoop chan bool, wg *sync.WaitGroup) {
	wg.Add(1)
	logger.Info("Starting process event loop", "PROCESS_EVENT_LOOP")
	// Load the default values according to the configuration
	processPoolingInterval = PopulateFromConfigs("harpy.resourceManager.eventLoop.processPoolingIntervalMS", processPoolingInterval, time.Millisecond)
	resourceRequestReleasedTimeout = PopulateFromConfigs("harpy.resourceManager.eventLoop.resourceRequestReleasedTimeoutSec", resourceRequestReleasedTimeout, time.Second)
	requestIdleTimeout = PopulateFromConfigs("harpy.resourceManager.eventLoop.requestIdleTimeoutSec", requestIdleTimeout, time.Second)

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
			EvalResourceRequestTimeout()
			EvalProviderTick(runningProvider)
			EvalNodeAllocation(runningProvider)
			EvalNodesReady()
			EvalResourceReleaseRequest()
			EvalRequestIdleTimeout()
			// TODO: We also need to implement logic to handle panics and nodes that are not able to come up
			EvalNodeShutdownCallbackProvider(runningProvider)
			time.Sleep(processPoolingInterval)
		}
	}
}

func ProcessEventLoop(runningProvider providers.ProviderInterface, existEventLoop chan bool, wg *sync.WaitGroup) error {
	go DoProcessLoop(runningProvider, existEventLoop, wg)
	return nil
}
