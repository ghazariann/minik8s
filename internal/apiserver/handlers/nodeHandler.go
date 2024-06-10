package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDNodePath, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, "Failed to fetch deployments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize a slice to hold the decoded deployment objects
	var NodeStores []apiobject.NodeStore

	// Iterate through each key-value pair returned from the store
	for _, kv := range resp.Kvs {
		var NodeStore apiobject.NodeStore
		if err := json.Unmarshal(kv.Value, &NodeStore); err != nil {
			http.Error(w, "Error decoding deployment data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		NodeStores = append(NodeStores, NodeStore)
	}

	// Convert the deployments slice to JSON
	NodeStoreJson, err := json.Marshal(NodeStores)
	if err != nil {
		http.Error(w, "Error encoding deployment data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Println("nodes fetched successfully")
	// Set content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(NodeStoreJson)

}
func AddNode(w http.ResponseWriter, r *http.Request) {
	var node apiobject.Node
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, _ := etcdclient.KeyExists(configs.ETCDNodePath + node.Metadata.Name)
	if res {
		http.Error(w, "Node already exists", http.StatusConflict)
		return
	}
	node.Metadata.UUID = uuid.New().String()

	nodeStore := node.ToStore()
	nodeStore.Status.Condition = "idle"
	nodeStore.Status.UpdateTime = time.Now()
	nodeStoreJson, err := json.Marshal(nodeStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO add namespace + name
	if err := etcdclient.PutKey(configs.ETCDNodePath+node.Metadata.Name, string(nodeStoreJson)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Node created: %s", node.Metadata.Name)
}
func GetNode(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is GET
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract node name from the query parameters
	nodeName := r.URL.Query().Get("name")
	if nodeName == "" {
		http.Error(w, "Node name is required", http.StatusBadRequest)
		return
	}

	// Retrieve node data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDNodePath+nodeName)
	if err != nil {
		http.Error(w, "Failed to fetch node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the node was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	// Unmarshal the node data
	var nodeStore apiobject.NodeStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &nodeStore); err != nil {
		http.Error(w, "Error decoding node data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the node data to JSON
	nodeStoreJson, err := json.Marshal(nodeStore)
	if err != nil {
		http.Error(w, "Error encoding node data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type and send the response

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(nodeStoreJson)
	// fmt.Fprintf(w, "Node fetched: %s", nodeName)
}
func UpdateNodeStatus(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is PUT

	// Extract node name from the query parameters
	nodeName := r.URL.Query().Get("name")
	if nodeName == "" {
		http.Error(w, "Node name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Node object
	var node apiobject.NodeStore
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing node data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDNodePath+nodeName)
	if err != nil {
		http.Error(w, "Failed to fetch node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the node was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing node data
	var nodeStore apiobject.NodeStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &nodeStore); err != nil {
		http.Error(w, "Error decoding node data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the node data ( status in running and has weave IP)
	nodeStore.Status = node.Status
	// Marshal the updated node data
	nodeStoreJson, err := json.Marshal(nodeStore)
	if err != nil {
		http.Error(w, "Error encoding node data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the node in etcd
	if err := etcdclient.PutKey(configs.ETCDNodePath+nodeName, string(nodeStoreJson)); err != nil {
		http.Error(w, "Failed to update node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "NodeStore updated: %s", nodeName)
}

func UpdateNode(w http.ResponseWriter, r *http.Request) {

	// Extract node name from the query parameters
	nodeName := r.URL.Query().Get("name")
	if nodeName == "" {
		http.Error(w, "Node name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Node object
	var node apiobject.NodeStore
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing node data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDNodePath+nodeName)
	if err != nil {
		http.Error(w, "Failed to fetch node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the node was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Node not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing node data
	var nodeStore apiobject.NodeStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &nodeStore); err != nil {
		http.Error(w, "Error decoding node data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the node data
	nodeStore.Spec = node.Spec
	nodeStore.Metadata.Labels = node.Metadata.Labels
	nodeStore.Status = node.Status
	// Marshal the updated node data
	nodeStoreJson, err := json.Marshal(nodeStore)
	if err != nil {
		http.Error(w, "Error encoding node data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the node in etcd
	if err := etcdclient.PutKey(configs.ETCDNodePath+nodeName, string(nodeStoreJson)); err != nil {
		http.Error(w, "Failed to update node: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "Node updated: %s", nodeName)
}

func DeleteNode(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is DELETE
	if r.Method != "DELETE" {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract node name from the query parameters
	nodeName := r.URL.Query().Get("name")
	if nodeName == "" {
		http.Error(w, "Node name is required", http.StatusBadRequest)
		return
	}

	// Delete the node from etcd
	nodeRes, err := etcdclient.GetKey(configs.ETCDNodePath + nodeName)

	if nodeRes == "" {
		http.Error(w, "Node "+nodeName+" does not exists: "+err.Error(), http.StatusInternalServerError)
	}
	node := apiobject.NodeStore{}
	err = json.Unmarshal([]byte(nodeRes), &node)
	if err != nil {
		http.Error(w, "Failed to decode node data: "+err.Error(), http.StatusInternalServerError)
	}
	err = etcdclient.DeleteKey(configs.ETCDNodePath + nodeName)

	if err != nil {
		http.Error(w, "Failed to delete node: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// delete endpoints
	for key, value := range node.Metadata.Labels {
		endpointsKVURL := path.Join(configs.ETCDEndpointPath, key, value, node.Metadata.UUID)
		etcdclient.DeleteKey(endpointsKVURL)
	}
	// Respond with confirmation
	fmt.Fprintf(w, "Node deleted: %s", nodeName)
}
