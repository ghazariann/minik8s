package handlers

import (
	"context"
	"encoding/json"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"net/http"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetEndpoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	// etcdclient.Cli.Delete(context.Background(), configs.ETCDHpaPath, clientv3.WithPrefix())
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDEndpointPath, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, "Failed to fetch deployments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize a slice to hold the decoded deployment objects
	var Endpoints []apiobject.Endpoint

	// Iterate through each key-value pair returned from the store
	for _, kv := range resp.Kvs {
		var endpoint apiobject.Endpoint
		if err := json.Unmarshal(kv.Value, &endpoint); err != nil {
			http.Error(w, "Error decoding deployment data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		Endpoints = append(Endpoints, endpoint)
	}

	// Convert the deployments slice to JSON
	HpaStoreJson, err := json.Marshal(Endpoints)
	if err != nil {
		http.Error(w, "Error encoding deployment data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Println("hpas fetched successfully")
	// Set content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(HpaStoreJson)

}
