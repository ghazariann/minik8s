package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"net/http"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func HandlePods(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		podName := r.URL.Query().Get("name")
		if podName == "" {
			http.Error(w, "Pod name is required", http.StatusBadRequest)
			return
		}
		podData, err := etcdclient.GetKey("pods/" + podName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Pod Data: %s", podData)
	case "POST":
		var pod apiobject.Pod
		if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		jsonData, err := json.Marshal(pod)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := etcdclient.PutKey("pods/"+pod.Name, string(jsonData)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Forward the pod information to kubelet for starting the pod
		// if err := forwardToKubelet(jsonData); err != nil {
		//   http.Error(w, "Failed to send pod start request to kubelet: "+err.Error(), http.StatusInternalServerError)
		//  return
		//}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Pod created: %s", pod.Name)
	default:
		http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
	}
}

func HandleAllPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	podsData, err := GetAllPods()
	if err != nil {
		http.Error(w, "Failed to fetch all pods: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(podsData)
}

func HandleUnscheduledPods(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	podsData, err := getUnscheduledPods()
	if err != nil {
		http.Error(w, "Failed to fetch unscheduled pods: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(podsData)
}

// handleUpdatePod 处理更新 Pod 状态和 NodeName 的请求
func HandleUpdatePod(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
		return
	}

	var updateRequest struct {
		Name     string `json:"name"`
		NodeName string `json:"nodeName"`
		Status   string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := updatePod(updateRequest.Name, updateRequest.NodeName, updateRequest.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Pod %s updated successfully", updateRequest.Name)
}

// updatePod 更新存储在 etcd 中的 Pod 状态和 NodeName
func updatePod(name, nodeName, status string) error {
	podKey := "pods/" + name
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := etcdclient.Cli.Get(ctx, podKey)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return fmt.Errorf("pod not found")
	}

	var pod apiobject.Pod
	if err := json.Unmarshal(resp.Kvs[0].Value, &pod); err != nil {
		return err
	}

	pod.NodeName = nodeName
	pod.Status = status

	jsonData, err := json.Marshal(pod)
	if err != nil {
		return err
	}

	_, err = etcdclient.Cli.Put(ctx, podKey, string(jsonData))
	return err
}

func GetAllPods() ([]byte, error) {
	return getPodsByCondition("")
}

func getUnscheduledPods() ([]byte, error) {
	return getPodsByCondition("NodeName")
}

func getPodsByCondition(filterKey string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := etcdclient.Cli.Get(ctx, "pods/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var podse []apiobject.Pod
	for _, kv := range resp.Kvs {
		var pod apiobject.Pod
		if err := json.Unmarshal(kv.Value, &pod); err == nil {
			if filterKey == "" || pod.NodeName == "" {
				podse = append(podse, pod)
			}
		}
	}
	return json.Marshal(podse)
}

func filterPodsBySelector(pods []apiobject.Pod, selector map[string]string) (selectedPods []apiobject.Pod) {
	for _, pod := range pods {
		matches := true
		for key, value := range selector {
			if pod.Labels[key] != value {
				matches = false
				break
			}
		}
		if matches {
			selectedPods = append(selectedPods, pod)
		}
	}
	return selectedPods
}
