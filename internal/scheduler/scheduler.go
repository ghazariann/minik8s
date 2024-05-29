package scheduler

import (
	//"context"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"minik8s/internal/apiobject"
	"net/http"
	"time"
)

var apiServerURL = "http://localhost:8080"

// StartScheduler starts the scheduler service
func StartScheduler() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		schedulePods()
	}
}

func schedulePods() {
	// 获取未调度的 Pods
	resp, err := http.Get(apiServerURL + "/unscheduled-pods")
	if err != nil {
		log.Printf("Error fetching unscheduled pods: %v", err)
		return
	}
	defer resp.Body.Close()

	var pods []apiobject.Pod
	if err := json.NewDecoder(resp.Body).Decode(&pods); err != nil {
		log.Printf("Error decoding unscheduled pods: %v", err)
		return
	}

	for _, pod := range pods {
		schedulePod(pod)
	}
}

func schedulePod(pod apiobject.Pod) {
	// 简单的轮询调度到第一个节点
	nodeURL := "http://localhost:10250" // 假设节点 URL 已知
	resp, err := http.Post(nodeURL+"/startPod", "application/json", jsonEncode(pod))
	if err != nil {
		log.Printf("Failed to schedule pod %s: %v", pod.Metadata.Name, err)
		return
	}
	defer resp.Body.Close()

	if err := updatePodStatusAndNode(pod.Metadata.Name, "nodeNameForExample", "Scheduled"); err != nil {
		log.Printf("Failed to update status for pod %s: %v", pod.Metadata.Name, err)
		return
	}

	log.Printf("Pod %s scheduled to %s", pod.Metadata.Name, nodeURL)
}

// updatePodStatusAndNode 向 API Server 发送更新 Pod 状态和节点名的请求
func updatePodStatusAndNode(podName, nodeName, status string) error {
	updateData := map[string]string{
		"name":     podName,
		"nodeName": nodeName,
		"status":   status,
	}

	data, err := json.Marshal(updateData)
	if err != nil {
		return fmt.Errorf("error marshaling update data: %v", err)
	}

	// 假设 API Server 的更新接口 URL 如下
	url := apiServerURL + "/updatePod"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending update request to API Server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API Server returned failure for update request: Status %d, Body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func jsonEncode(pod apiobject.Pod) *bytes.Buffer {
	data, err := json.Marshal(pod)
	if err != nil {
		log.Printf("Error encoding pod: %v", err)
		return nil
	}
	return bytes.NewBuffer(data)
}
