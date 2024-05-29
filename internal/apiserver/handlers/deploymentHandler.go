package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"net/http"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetDeployments(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Assuming 'etcdclient' is an initialized client that can interact with etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDDeploymentPath, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, "Failed to fetch deployments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize a slice to hold the decoded deployment objects
	var deploymentStores []apiobject.DeploymentStore

	// Iterate through each key-value pair returned from the store
	for _, kv := range resp.Kvs {
		var deploymentStore apiobject.DeploymentStore
		if err := json.Unmarshal(kv.Value, &deploymentStore); err != nil {
			http.Error(w, "Error decoding deployment data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		deploymentStores = append(deploymentStores, deploymentStore)
	}

	// Convert the deployments slice to JSON
	deploymentsJSON, err := json.Marshal(deploymentStores)
	if err != nil {
		http.Error(w, "Error encoding deployment data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(deploymentsJSON)
}

func GetDeployment(w http.ResponseWriter, r *http.Request) {
	deploymentName := r.URL.Query().Get("name")
	if deploymentName == "" {
		http.Error(w, "Deployment name is required", http.StatusBadRequest)
		return
	}
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDDeploymentPath+deploymentName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(resp.Kvs) == 0 {
		http.Error(w, "Deployment not found", http.StatusNotFound)
		return
	}
	// Unmarshal the pod data
	var deploymentStore apiobject.DeploymentStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &deploymentStore); err != nil {
		http.Error(w, "Error decoding deployment data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the pod data to JSON
	deploymentStoreJson, err := json.Marshal(deploymentStore)
	if err != nil {
		http.Error(w, "Error encoding pod data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type and send the response

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(deploymentStoreJson)
}

func AddDeployment(w http.ResponseWriter, r *http.Request) {
	var deployment apiobject.Deployment
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, _ := etcdclient.GetKey(configs.ETCDDeploymentPath + deployment.Metadata.Name)
	if res != "" {
		http.Error(w, "Deployment already exists", http.StatusBadRequest)
		return
	}
	deployment.Metadata.UUID = uuid.New().String()

	deploymentStore := deployment.ToDeploymentStore()

	jsonData, err := json.Marshal(deploymentStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := etcdclient.PutKey(configs.ETCDDeploymentPath+deployment.Metadata.Name, string(jsonData)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Deployment created: %s", deployment.Metadata.Name)
}

func DeleteDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	deploymentName := r.URL.Query().Get("name")
	if deploymentName == "" {
		http.Error(w, "Deployment name is required", http.StatusBadRequest)
		return
	}
	// etcdclient.Cli.Get(context.Background(), configs.ETCDDeploymentPath+deploymentName)
	if _, err := etcdclient.Cli.Delete(context.Background(), configs.ETCDDeploymentPath+deploymentName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// res, _ := etcdclient.GetKey(configs.ETCDDeploymentPath + deploymentName)
	// if res != "" {
	// 	http.Error(w, "Deployment not deleted", http.StatusInternalServerError)
	// 	return
	// }
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deployment deleted: %s", deploymentName)
}

func UpdateDeployment(w http.ResponseWriter, r *http.Request) {
	var deployment apiobject.DeploymentStore
	if err := json.NewDecoder(r.Body).Decode(&deployment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// deploymentStore := deployment.ToDeploymentStore()

	jsonData, err := json.Marshal(deployment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := etcdclient.PutKey(configs.ETCDDeploymentPath+deployment.Metadata.Name, string(jsonData)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deployment updated: %s", deployment.Metadata.Name)
}
