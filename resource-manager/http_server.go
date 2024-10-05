// Description: This file contains the http server implementation for the resource manager.
package main

import (
	"encoding/json"
	"net"
	"net/http"
	"resource-manager/logger"
	"strings"
	"sync"
	"time"
)

// Resource Manager HTTP Server
type ResourceManagerHTTPServer struct {
}

type HealthCheckResponse struct {
	Status string `json:"status"`
}

type Node struct {
	NodeID                string `json:"node_id"`
	NodeType              string `json:"node_type"`
	NodeGRPCAddress       string `json:"node_grpc_address"`
	NodeStatus            string `json:"node_status"`
	LastHeartbeatReceived string `json:"last_heartbeat_received"`
	IsServingRequest      bool   `json:"is_serving_request"`
	ServingRequestID      string `json:"serving_request_id"`
}

type Request struct {
	RequestID             string `json:"request_id"`
	NodeType              string `json:"node_type"`
	NodeCount             uint32 `json:"node_count"`
	ServingStatus         string `json:"serving_status"`
	LastHeartbeatReceived string `json:"last_heartbeat_received"`
}

func toIsoDate(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(time.RFC3339)
}

// NewResourceManagerHTTPServer creates a new instance of the resource manager http server
func NewResourceManagerHTTPServer() *ResourceManagerHTTPServer {
	return &ResourceManagerHTTPServer{}
}

func WriteJSONResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		logger.Error("Failed to write JSON response", "HTTP_SERVER", err)
		http.Error(w, "Failed to write JSON response", http.StatusInternalServerError)
	}
}

// HealthCheckHandler is the handler for the health check endpoint
func (s *ResourceManagerHTTPServer) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check is the request is a GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create the response
	response := HealthCheckResponse{
		Status: "OK",
	}

	// Write the response (JSON)
	WriteJSONResponse(w, response)
}

func (s *ResourceManagerHTTPServer) NodesHandler(w http.ResponseWriter, r *http.Request) {
	// Check is the request is a GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Check for filter parameters
	nodeID := r.URL.Query().Get("node_id")
	nodeType := r.URL.Query().Get("node_type")
	nodeStatus := StringToEnum(TypeMarker_LiveNodeStatusEnum, r.URL.Query().Get("node_status"))

	// Get the nodes from the database
	queryComponents := []string{}
	queryParams := []interface{}{}
	if nodeID != "" {
		queryComponents = append(queryComponents, "node_id = ?")
		queryParams = append(queryParams, nodeID)
	}
	if nodeType != "" {
		queryComponents = append(queryComponents, "node_type = ?")
		queryParams = append(queryParams, nodeType)
	}
	if nodeStatus != nil {
		queryComponents = append(queryComponents, "node_status = ?")
		queryParams = append(queryParams, nodeStatus)
	}
	var results []*LiveNode
	if len(queryComponents) == 0 {
		results = LiveNodes()
	} else {
		where := strings.Join(queryComponents, " AND ")
		results = LiveNodesWithWhereClause(where, queryParams)
	}

	// Create the response
	nodes := []Node{}
	for _, node := range results {
		nodes = append(nodes, Node{
			NodeID:                node.NodeID,
			NodeType:              node.NodeType,
			NodeGRPCAddress:       node.NodeGRPCAddress,
			NodeStatus:            EnumToString(node.NodeStatus),
			LastHeartbeatReceived: toIsoDate(node.LastHeartbeatReceived),
			IsServingRequest:      node.IsServingRequest,
			ServingRequestID:      node.ServingRequestID,
		})
	}

	// Write the response (JSON)
	WriteJSONResponse(w, nodes)
}

func (s *ResourceManagerHTTPServer) RequestsHandler(w http.ResponseWriter, r *http.Request) {
	// Check is the request is a GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Check for filter parameters
	requestID := r.URL.Query().Get("request_id")
	nodeType := r.URL.Query().Get("node_type")
	servingStatus := StringToEnum(TypeMarker_ResourceRequestStatusEnum, r.URL.Query().Get("serving_status"))

	queryComponents := []string{}
	queryParams := []interface{}{}
	if requestID != "" {
		queryComponents = append(queryComponents, "request_id = ?")
		queryParams = append(queryParams, requestID)
	}
	if nodeType != "" {
		queryComponents = append(queryComponents, "node_type = ?")
		queryParams = append(queryParams, nodeType)
	}
	if servingStatus != nil {
		queryComponents = append(queryComponents, "serving_status = ?")
		queryParams = append(queryParams, servingStatus)
	}

	// Get the requests from the database
	var results []*ResourceAssignment
	if len(queryComponents) == 0 {
		results = ResourceAssignments()
	} else {
		where := strings.Join(queryComponents, " AND ")
		results = ResourceAssignmentsWithWhereClause(where, queryParams)
	}

	// Create the response
	requests := []Request{}
	for _, request := range results {
		requests = append(requests, Request{
			RequestID:             request.RequestID,
			NodeType:              request.NodeType,
			NodeCount:             request.NodeCount,
			ServingStatus:         EnumToString(request.ServingStatus),
			LastHeartbeatReceived: toIsoDate(request.LastHeartbeatReceived),
		})
	}

	// Write the response (JSON)
	WriteJSONResponse(w, requests)
}

func (s *ResourceManagerHTTPServer) CatalogHandler(w http.ResponseWriter, r *http.Request) {
	// Check is the request is a GET request
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for filter parameters
	nodeType := r.URL.Query().Get("node_type")

	if nodeType != "" {
		// Get the catalog from the database
		result, ok := GetNodeCatalog(nodeType)
		if !ok {
			http.Error(w, "Catalog not found", http.StatusNotFound)
			return
		}

		// Write the response (JSON)
		WriteJSONResponse(w, result)
		return
	}

	// Get the catalog from the database
	results := NodeCatalogs()

	// Write the response (JSON)
	WriteJSONResponse(w, results)
}

func (s *ResourceManagerHTTPServer) ConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Return not implemented status
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// StartServer starts the http server
func StartServer(stopChan chan bool, wg *sync.WaitGroup) error {
	// The server serves to expose the current state of the deployment of the Harpy system
	// The server will expose the following endpoints:
	// /health - This endpoint will return a 200 OK if the server is up
	// /nodes - This endpoint will return the current state of the nodes
	// /requests - This endpoint will return the current state of the requests
	// /config - This endpoint will return the current configuration of the resource manager
	// /catalog - This endpoint will return the current catalog of the nodes

	logger.Info("Starting HTTP server", "HTTP_SERVER")

	// Create the server
	s := NewResourceManagerHTTPServer()

	// Register the handlers
	http.HandleFunc("/health", s.HealthCheckHandler)
	http.HandleFunc("/nodes", s.NodesHandler)
	http.HandleFunc("/requests", s.RequestsHandler)
	http.HandleFunc("/config", s.ConfigHandler)
	http.HandleFunc("/catalog", s.CatalogHandler)

	// Start the server - This is a blocking call and we need to be able to stop the server gracefully
	// when the stopChan is closed
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Error("Failed to start HTTP server", "HTTP_SERVER", err)
		return err
	}

	go func() {
		if err := http.Serve(l, nil); err != nil {
			logger.Error("HTTP server stopped with error", "HTTP_SERVER", err)
		}
	}()

	go func() {
		<-stopChan
		logger.Info("Stopping HTTP server", "HTTP_SERVER")
		l.Close()
		wg.Done()
	}()

	return nil
}
