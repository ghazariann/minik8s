package handlers

import (
	"encoding/json"
	"fmt"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/endpoints"
	"net/http"
)

func HandleServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var service apiobject.Service
		if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
			http.Error(w, "Invalid service data", http.StatusBadRequest)
			return
		}

		// 将服务数据转换为 JSON 并存储
		serviceData, err := json.Marshal(service)
		if err != nil {
			http.Error(w, "Failed to encode service data", http.StatusInternalServerError)
			return
		}
		serviceKey := "services/" + service.Name
		if err := etcdclient.PutKey(serviceKey, string(serviceData)); err != nil {
			http.Error(w, "Failed to store service in etcd", http.StatusInternalServerError)
			return
		}

		// 获取所有Pods并筛选符合Service选择器的Pods
		podsData, err := GetAllPods()
		if err != nil {
			http.Error(w, "Failed to fetch pods for service endpoints", http.StatusInternalServerError)
			return
		}
		var allPods []apiobject.Pod
		json.Unmarshal(podsData, &allPods)
		matchedPods := filterPodsBySelector(allPods, service.Selector)

		// 创建并存储对应的 Endpoint 对象
		var ep endpoints.Endpoint
		ep.ServiceName = service.Name
		for _, pod := range matchedPods {
			ep.IPs = append(ep.IPs, pod.Metadata.UUID) // WRONG !! TODO: Get pod IP
		}
		epData, err := json.Marshal(ep)
		if err != nil {
			http.Error(w, "Failed to encode endpoint data", http.StatusInternalServerError)
			return
		}
		epKey := "endpoints/" + service.Name
		if err := etcdclient.PutKey(epKey, string(epData)); err != nil {
			http.Error(w, "Failed to store endpoint in etcd", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Service and endpoint created successfully")

	case "GET":
		serviceName := r.URL.Query().Get("name")
		if serviceName == "" {
			http.Error(w, "Service name is required", http.StatusBadRequest)
			return
		}
		serviceData, err := etcdclient.GetKey("services/" + serviceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Service Data: %s", serviceData)

	}
}

func HandleAllServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	servicesData, err := etcdclient.GetAllKeys("services/")
	if err != nil {
		http.Error(w, "Failed to fetch services: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(servicesData)
}
