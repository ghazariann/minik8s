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

	"github.com/google/uuid"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetHpas(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	// etcdclient.Cli.Delete(context.Background(), configs.ETCDHpaPath, clientv3.WithPrefix())
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDHpaPath, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, "Failed to fetch deployments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize a slice to hold the decoded deployment objects
	var HpaStores []apiobject.HpaStore

	// Iterate through each key-value pair returned from the store
	for _, kv := range resp.Kvs {
		var HpaStore apiobject.HpaStore
		if err := json.Unmarshal(kv.Value, &HpaStore); err != nil {
			http.Error(w, "Error decoding deployment data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		HpaStores = append(HpaStores, HpaStore)
	}

	// Convert the deployments slice to JSON
	HpaStoreJson, err := json.Marshal(HpaStores)
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
func AddHpa(w http.ResponseWriter, r *http.Request) {
	var hpa apiobject.Hpa
	if err := json.NewDecoder(r.Body).Decode(&hpa); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, _ := etcdclient.KeyExists(configs.ETCDHpaPath + hpa.Metadata.Name)
	if res {
		http.Error(w, "Hpa already exists", http.StatusConflict)
		return
	}
	hpa.Metadata.UUID = uuid.New().String()

	hpaStore := hpa.ToStore()

	hpaStoreJson, err := json.Marshal(hpaStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO add namespace + name
	if err := etcdclient.PutKey(configs.ETCDHpaPath+hpa.Metadata.Name, string(hpaStoreJson)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Hpa created: %s", hpa.Metadata.Name)
}
func GetHpa(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is GET
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract hpa name from the query parameters
	hpaName := r.URL.Query().Get("name")
	if hpaName == "" {
		http.Error(w, "Hpa name is required", http.StatusBadRequest)
		return
	}

	// Retrieve hpa data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDHpaPath+hpaName)
	if err != nil {
		http.Error(w, "Failed to fetch hpa: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the hpa was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Hpa not found", http.StatusNotFound)
		return
	}

	// Unmarshal the hpa data
	var hpaStore apiobject.HpaStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &hpaStore); err != nil {
		http.Error(w, "Error decoding hpa data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the hpa data to JSON
	hpaStoreJson, err := json.Marshal(hpaStore)
	if err != nil {
		http.Error(w, "Error encoding hpa data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type and send the response

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(hpaStoreJson)
	// fmt.Fprintf(w, "Hpa fetched: %s", hpaName)
}
func UpdateHpaStatus(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is PUT

	// Extract hpa name from the query parameters
	hpaName := r.URL.Query().Get("name")
	if hpaName == "" {
		http.Error(w, "Hpa name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Hpa object
	var hpa apiobject.HpaStore
	if err := json.NewDecoder(r.Body).Decode(&hpa); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing hpa data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDHpaPath+hpaName)
	if err != nil {
		http.Error(w, "Failed to fetch hpa: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the hpa was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Hpa not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing hpa data
	var hpaStore apiobject.HpaStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &hpaStore); err != nil {
		http.Error(w, "Error decoding hpa data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the hpa data ( status in running and has weave IP)
	hpaStore.Status = hpa.Status
	// Marshal the updated hpa data
	hpaStoreJson, err := json.Marshal(hpaStore)
	if err != nil {
		http.Error(w, "Error encoding hpa data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the hpa in etcd
	if err := etcdclient.PutKey(configs.ETCDHpaPath+hpaName, string(hpaStoreJson)); err != nil {
		http.Error(w, "Failed to update hpa: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "HpaStore updated: %s", hpaName)
}

func UpdateHpa(w http.ResponseWriter, r *http.Request) {

	// Extract hpa name from the query parameters
	hpaName := r.URL.Query().Get("name")
	if hpaName == "" {
		http.Error(w, "Hpa name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Hpa object
	var hpa apiobject.HpaStore
	if err := json.NewDecoder(r.Body).Decode(&hpa); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing hpa data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDHpaPath+hpaName)
	if err != nil {
		http.Error(w, "Failed to fetch hpa: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the hpa was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Hpa not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing hpa data
	var hpaStore apiobject.HpaStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &hpaStore); err != nil {
		http.Error(w, "Error decoding hpa data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the hpa data
	hpaStore.Spec = hpa.Spec
	hpaStore.Metadata.Labels = hpa.Metadata.Labels
	hpaStore.Status = hpa.Status
	// Marshal the updated hpa data
	hpaStoreJson, err := json.Marshal(hpaStore)
	if err != nil {
		http.Error(w, "Error encoding hpa data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the hpa in etcd
	if err := etcdclient.PutKey(configs.ETCDHpaPath+hpaName, string(hpaStoreJson)); err != nil {
		http.Error(w, "Failed to update hpa: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "Hpa updated: %s", hpaName)
}

func DeleteHpa(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is DELETE
	if r.Method != "DELETE" {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract hpa name from the query parameters
	hpaName := r.URL.Query().Get("name")
	if hpaName == "" {
		http.Error(w, "Hpa name is required", http.StatusBadRequest)
		return
	}

	// Delete the hpa from etcd
	hpaRes, err := etcdclient.GetKey(configs.ETCDHpaPath + hpaName)

	if hpaRes == "" {
		http.Error(w, "Hpa "+hpaName+" does not exists: "+err.Error(), http.StatusInternalServerError)
	}
	hpa := apiobject.HpaStore{}
	err = json.Unmarshal([]byte(hpaRes), &hpa)
	if err != nil {
		http.Error(w, "Failed to decode hpa data: "+err.Error(), http.StatusInternalServerError)
	}
	err = etcdclient.DeleteKey(configs.ETCDHpaPath + hpaName)

	if err != nil {
		http.Error(w, "Failed to delete hpa: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// delete endpoints
	for key, value := range hpa.Metadata.Labels {
		endpointsKVURL := path.Join(configs.ETCDEndpointPath, key, value, hpa.Metadata.UUID)
		etcdclient.DeleteKey(endpointsKVURL)
	}
	// Respond with confirmation
	fmt.Fprintf(w, "Hpa deleted: %s", hpaName)
}
