package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/apiserver/helpers"
	"minik8s/internal/configs"
	"minik8s/internal/scheduler"
	"net/http"
	"path"
	"time"

	"github.com/google/uuid"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	// etcdclient.Cli.Delete(context.Background(), configs.ETCDPodPath, clientv3.WithPrefix())
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDPodPath, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, "Failed to fetch deployments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize a slice to hold the decoded deployment objects
	var PodStores []apiobject.PodStore

	// Iterate through each key-value pair returned from the store
	for _, kv := range resp.Kvs {
		var PodStore apiobject.PodStore
		if err := json.Unmarshal(kv.Value, &PodStore); err != nil {
			http.Error(w, "Error decoding deployment data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		PodStores = append(PodStores, PodStore)
	}

	// Convert the deployments slice to JSON
	PodStoreJson, err := json.Marshal(PodStores)
	if err != nil {
		http.Error(w, "Error encoding deployment data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Println("pods fetched successfully")
	// Set content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(PodStoreJson)

}
func AddPod(w http.ResponseWriter, r *http.Request) {
	var pod apiobject.Pod
	if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, _ := etcdclient.KeyExists(configs.ETCDPodPath + pod.Metadata.Name)
	if res {
		http.Error(w, "Pod already exists", http.StatusConflict)
		return
	}
	pod.Metadata.UUID = uuid.New().String()

	podStore := pod.ToStore()

	podStore.Status.Phase = apiobject.PodPending
	podStore.Status.LastUpdated = time.Now()
	// TODO add namespace + name
	shceduledNode, err := scheduler.SchedulePod(podStore)

	if err != nil {
		log.Fatal("Failed to schedule pod")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Schedule to node: ", shceduledNode)
	podStoreJson, err := json.Marshal(podStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := etcdclient.PutKey(configs.ETCDPodPath+pod.Metadata.Name, string(podStoreJson)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Pod created: %s", pod.Metadata.Name)
}
func GetPod(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is GET
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract pod name from the query parameters
	podName := r.URL.Query().Get("name")
	if podName == "" {
		http.Error(w, "Pod name is required", http.StatusBadRequest)
		return
	}

	// Retrieve pod data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDPodPath+podName)
	if err != nil {
		http.Error(w, "Failed to fetch pod: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the pod was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Pod not found", http.StatusNotFound)
		return
	}

	// Unmarshal the pod data
	var podStore apiobject.PodStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &podStore); err != nil {
		http.Error(w, "Error decoding pod data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the pod data to JSON
	podStoreJson, err := json.Marshal(podStore)
	if err != nil {
		http.Error(w, "Error encoding pod data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(podStoreJson)
	// fmt.Fprintf(w, "Pod fetched: %s", podName)
}
func UpdatePodStatus(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is PUT

	// Extract pod name from the query parameters
	podName := r.URL.Query().Get("name")
	if podName == "" {
		http.Error(w, "Pod name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Pod object
	var pod apiobject.PodStore
	if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing pod data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDPodPath+podName)
	if err != nil {
		http.Error(w, "Failed to fetch pod: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the pod was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Pod not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing pod data
	var podStore apiobject.PodStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &podStore); err != nil {
		http.Error(w, "Error decoding pod data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the pod data ( status in running and has weave IP)
	podStore.Status = pod.Status
	helpers.UpdateEndPoints(&podStore) // will update service too
	podStore.Spec.NodeName = pod.Spec.NodeName
	podStore.Status.LastUpdated = time.Now()
	// Marshal the updated pod data
	podStoreJson, err := json.Marshal(podStore)
	if err != nil {
		http.Error(w, "Error encoding pod data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the pod in etcd
	if err := etcdclient.PutKey(configs.ETCDPodPath+podName, string(podStoreJson)); err != nil {
		http.Error(w, "Failed to update pod: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "PodStore updated: %s", podName)
}

func UpdatePod(w http.ResponseWriter, r *http.Request) {

	// Extract pod name from the query parameters
	podName := r.URL.Query().Get("name")
	if podName == "" {
		http.Error(w, "Pod name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Pod object
	var pod apiobject.Pod
	if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing pod data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDPodPath+podName)
	if err != nil {
		http.Error(w, "Failed to fetch pod: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the pod was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Pod not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing pod data
	var podStore apiobject.PodStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &podStore); err != nil {
		http.Error(w, "Error decoding pod data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the pod data
	podStore.Spec = pod.Spec
	podStore.Metadata.Labels = pod.Metadata.Labels
	podStore.Status.LastUpdated = time.Now()

	// Marshal the updated pod data
	podStoreJson, err := json.Marshal(podStore)
	if err != nil {
		http.Error(w, "Error encoding pod data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the pod in etcd
	if err := etcdclient.PutKey(configs.ETCDPodPath+podName, string(podStoreJson)); err != nil {
		http.Error(w, "Failed to update pod: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "Pod updated: %s", podName)
}

func DeletePod(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is DELETE
	if r.Method != "DELETE" {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract pod name from the query parameters
	podName := r.URL.Query().Get("name")
	if podName == "" {
		http.Error(w, "Pod name is required", http.StatusBadRequest)
		return
	}

	// Delete the pod from etcd
	podRes, err := etcdclient.GetKey(configs.ETCDPodPath + podName)

	if podRes == "" {
		http.Error(w, "Pod "+podName+" does not exists: "+err.Error(), http.StatusInternalServerError)
	}
	pod := apiobject.PodStore{}
	err = json.Unmarshal([]byte(podRes), &pod)
	if err != nil {
		http.Error(w, "Failed to decode pod data: "+err.Error(), http.StatusInternalServerError)
	}
	err = etcdclient.DeleteKey(configs.ETCDPodPath + podName)

	if err != nil {
		http.Error(w, "Failed to delete pod: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// delete endpoints
	for key, value := range pod.Metadata.Labels {
		endpointsKVURL := path.Join(configs.ETCDEndpointPath, key, value, pod.Metadata.UUID)
		etcdclient.DeleteKey(endpointsKVURL)
		// update service

	}
	// Respond with confirmation
	fmt.Fprintf(w, "Pod deleted: %s", podName)
}
