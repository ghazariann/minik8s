package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"net/http"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// GetServices fetches all services from etcd
func GetServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	resp, err := etcdclient.Cli.Get(context.Background(), "services/", clientv3.WithPrefix())
	if err != nil {
		http.Error(w, "Failed to fetch services: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var services []apiobject.ServiceStore

	for _, kv := range resp.Kvs {
		var service apiobject.ServiceStore
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			http.Error(w, "Error decoding service data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		services = append(services, service)
	}

	servicesJSON, err := json.Marshal(services)
	if err != nil {
		http.Error(w, "Error encoding service data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(servicesJSON)
}

// AddService adds a new service to etcd
func AddService(w http.ResponseWriter, r *http.Request) {
	var service apiobject.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if exists, _ := etcdclient.KeyExists("services/" + service.Metadata.Name); exists {
		http.Error(w, "Service already exists", http.StatusConflict)
		return
	}

	service.Metadata.UUID = uuid.New().String()
	serviceStore := service.ToServiceStore()

	serviceStoreJSON, err := json.Marshal(serviceStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := etcdclient.PutKey("services/"+service.Metadata.Name, string(serviceStoreJSON)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Service created: %s", service.Metadata.Name)
}

// GetService fetches a specific service from etcd by name
func GetService(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("name")
	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	resp, err := etcdclient.Cli.Get(context.Background(), "services/"+serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(resp.Kvs) == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	var serviceStore apiobject.ServiceStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &serviceStore); err != nil {
		http.Error(w, "Error decoding service data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	serviceStoreJSON, err := json.Marshal(serviceStore)
	if err != nil {
		http.Error(w, "Error encoding service data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(serviceStoreJSON)
	fmt.Fprintf(w, "Service fetched: %s", serviceName)
}

// DeleteService deletes a specific service from etcd by name
func DeleteService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("name")
	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	if _, err := etcdclient.Cli.Delete(context.Background(), "services/"+serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Service deleted: %s", serviceName)
}

// UpdateService updates an existing service in etcd
func UpdateService(w http.ResponseWriter, r *http.Request) {
	var service apiobject.Service
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	serviceStore := service.ToServiceStore()

	serviceStoreJSON, err := json.Marshal(serviceStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := etcdclient.PutKey("services/"+service.Metadata.Name, string(serviceStoreJSON)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Service updated: %s", service.Metadata.Name)
}
