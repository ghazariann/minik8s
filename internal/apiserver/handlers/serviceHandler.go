package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"net/http"
	"strconv"

	"minik8s/internal/apiserver/helpers"

	"path"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// GetServices fetches all services from etcd
func GetServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDServicePath, clientv3.WithPrefix())
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

	if exists, _ := etcdclient.KeyExists(configs.ETCDServicePath + service.Metadata.Name); exists {
		http.Error(w, "Service already exists", http.StatusConflict)
		return
	}
	service.Spec.ClusterIP, _ = helpers.AllocateNewClusterIP()
	service.Metadata.UUID = uuid.New().String()
	serviceStore := service.ToServiceStore()
	serviceStore.Status.Phase = "pending"
	serviceStoreJSON, err := json.Marshal(serviceStore)

	for key, value := range service.Spec.Selector {
		func(key, value string) {
			svcSelectorURL := path.Join(configs.ETCDServiceSelectorPath, key, value, service.Metadata.UUID)

			if err := etcdclient.PutKey(svcSelectorURL, string(serviceStoreJSON)); err != nil {
				return
			}
			var endpoints []apiobject.Endpoint
			if endpoints, err = helpers.GetEndpoints(key, value); err != nil {
				return
			} else {
				serviceStore.Status.Endpoints = append(serviceStore.Status.Endpoints, endpoints...)
			}

			log.Printf("APIServer: endpoints number of endpoints in service " + service.Metadata.Name + " is " + strconv.Itoa(len(serviceStore.Status.Endpoints)))

		}(key, value)
	}
	serviceStoreJSON, err = json.Marshal(serviceStore) // added endpoints
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := etcdclient.PutKey(configs.ETCDServicePath+service.Metadata.Name, string(serviceStoreJSON)); err != nil {
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

	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDServicePath+serviceName)
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
	// fmt.Fprintf(w, "Service fetched: %s", serviceName)
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
	serviceResp, _ := etcdclient.Cli.Get(context.Background(), configs.ETCDServicePath+serviceName)
	if len(serviceResp.Kvs) == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}
	service := apiobject.ServiceStore{}
	if err := json.Unmarshal(serviceResp.Kvs[0].Value, &service); err != nil {
		http.Error(w, "Error decoding service data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := etcdclient.Cli.Delete(context.Background(), configs.ETCDServicePath+serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for key, value := range service.Spec.Selector {
		func(key, value string) {
			svcSelectorURL := path.Join(configs.ETCDServiceSelectorPath, key, value, service.Metadata.UUID)

			if err := etcdclient.DeleteKey(svcSelectorURL); err != nil {
				return
			}

		}(key, value)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Service deleted: %s", serviceName)
}

func UpdateServiceStatus(w http.ResponseWriter, r *http.Request) error {
	var service apiobject.ServiceStore
	if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
		return err
	}

	serviceStoreJSON, err := json.Marshal(service)
	if err != nil {
		return err
	}
	if err := etcdclient.PutKey(configs.ETCDServicePath+service.Metadata.Name, string(serviceStoreJSON)); err != nil {
		return err
	}

	return nil
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

	if err := etcdclient.PutKey(configs.ETCDServicePath+service.Metadata.Name, string(serviceStoreJSON)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Service updated: %s", service.Metadata.Name)
}
