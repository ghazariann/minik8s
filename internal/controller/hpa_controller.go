package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"net/http"
	"time"
)

func CalcAvgCpuPercentage(pods []apiobject.PodStore) float64 {
	totalCpu := 0.0
	for _, pod := range pods {
		totalCpu += pod.Status.CpuPercent
	}
	avgCpu := totalCpu / float64(len(pods))
	return avgCpu * 100
}

func CalcAvgMemPercentage(pods []apiobject.PodStore) float64 {
	totalMem := 0.0
	for _, pod := range pods {
		totalMem += pod.Status.MemPercent
	}
	avgMem := totalMem / float64(len(pods))
	return avgMem * 100
}

func CalculatedesiredMetricValue(hpa apiobject.HpaStore, cpuUsage float64, memoryUsage float64) int {
	// >1 means resource is not enough
	desiredMetricValue := int(math.Ceil(math.Max(cpuUsage/float64(hpa.Spec.Metrics.CPUPercent),
		memoryUsage/float64(hpa.Spec.Metrics.MemPercent)) * float64(hpa.Status.CurrentReplicas)))
	// adjust the expected pod number to be within the min and max replicas
	if desiredMetricValue < hpa.Spec.MinReplicas {
		desiredMetricValue = hpa.Spec.MinReplicas
	}
	if desiredMetricValue > hpa.Spec.MaxReplicas {
		desiredMetricValue = hpa.Spec.MaxReplicas
	}
	return desiredMetricValue
}

func GetHpasFromAPIServer() ([]apiobject.HpaStore, error) {
	url := configs.GetApiServerUrl() + configs.HpasUrl

	allHpas := make([]apiobject.HpaStore, 0)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error sending request to list hpas: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if json.Unmarshal(body, &allHpas) != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}
	return allHpas, nil
}

func AddHpaPod(hpaMeta *apiobject.Metadata, pod *apiobject.PodStore) error {
	// replicate the same pod with added random string prefix for the name and container names
	url := configs.GetApiServerUrl() + configs.PodsURL
	newPod := pod
	newPod.Metadata.Name = pod.Metadata.Name + "-" + RandomStr(5)

	for index := range pod.Spec.Containers {
		newPod.Spec.Containers[index].Name = pod.Spec.Containers[index].Name + "-" + RandomStr(5)
	}
	jsonData, _ := json.Marshal(newPod)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		return err
	}

	return nil
}

func ReduceHpaPod(pod apiobject.PodStore) error {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.PodUrl+"?name=%s", pod.Metadata.Name)
	req, _ := http.NewRequest("DELETE", url, nil)
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}

func UpdateHpaStatus(hpa apiobject.HpaStore) error {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.HpaUrl+"?name=%s", hpa.Metadata.Name)
	jsonData, _ := json.Marshal(hpa)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	return nil
}

func HpaUpdate(hpa apiobject.HpaStore, pods []apiobject.PodStore) {
	filteredPods := make([]apiobject.PodStore, 0)
	for _, pod := range pods {
		if FilterBySelector(&pod, hpa.Spec.Selector.MatchLabels) {
			filteredPods = append(filteredPods, pod)
		}
	}
	if len(filteredPods) == 0 {
		return
	}
	hpa.Status.CurrentReplicas = len(filteredPods)
	// Continuously adjust HPA until desired conditions are met
	for {
		avgCPU := CalcAvgCpuPercentage(filteredPods)
		avgMem := CalcAvgMemPercentage(filteredPods)

		desiredMetricValue := CalculatedesiredMetricValue(hpa, avgCPU, avgMem)
		hpa.Status.CurrentCPUPercent = avgCPU
		hpa.Status.CurrentMemPercent = avgMem
		if hpa.Status.CurrentReplicas < hpa.Spec.MinReplicas {
			AddHpaPod(&hpa.Metadata, &filteredPods[0])
			hpa.Status.CurrentReplicas++
		} else if hpa.Status.CurrentReplicas > hpa.Spec.MaxReplicas {
			ReduceHpaPod(filteredPods[len(filteredPods)-1])
			hpa.Status.CurrentReplicas--
			filteredPods = filteredPods[:len(filteredPods)-1]
		} else if desiredMetricValue > hpa.Status.CurrentReplicas {
			AddHpaPod(&hpa.Metadata, &filteredPods[0])
			hpa.Status.CurrentReplicas++
		} else if desiredMetricValue < hpa.Status.CurrentReplicas {
			ReduceHpaPod(filteredPods[len(filteredPods)-1])
			hpa.Status.CurrentReplicas--
			filteredPods = filteredPods[:len(filteredPods)-1]
		} else {
			break // Break the loop if no adjustment is needed
		}

		// Update status after making changes
		err := UpdateHpaStatus(hpa)
		if err != nil {
			log.Printf("Error updating HPA status: %v", err)
		}

		fmt.Printf("Sleeping 15 seconds")
		time.Sleep(time.Duration(15 * time.Second))
	}
}

func HpaRoutine() {
	pods, err := GetPodsFromAPIServer()
	if err != nil {
		return
	}
	currentHpas, err := GetHpasFromAPIServer()

	if err != nil {
		return
	}
	// Detect deleted hpas and delete corresponding pods
	for _, prevHpa := range previousHpas {
		found := false
		for _, currHpa := range currentHpas {
			if prevHpa.Metadata.UUID == currHpa.Metadata.UUID {
				found = true
				break
			}
		}
		if !found {
			filteredPodsWithoutHpa := make([]apiobject.PodStore, 0)
			for _, pod := range pods {
				if FilterBySelector(&pod, prevHpa.Spec.Selector.MatchLabels) {
					filteredPodsWithoutHpa = append(filteredPodsWithoutHpa, pod)
				}
			}
			for len(filteredPodsWithoutHpa) > 1 {
				// Get the index of the last pod in the slice
				lastPodIndex := len(filteredPodsWithoutHpa) - 1

				// Reduce resources or handle the pod specified by lastPodIndex
				ReduceHpaPod(filteredPodsWithoutHpa[lastPodIndex])

				// Remove the last pod from the slice
				filteredPodsWithoutHpa = filteredPodsWithoutHpa[:lastPodIndex]
			}
		}
	}

	// Update previous deployments for the next iteration
	previousHpas = currentHpas

	for _, hpa := range currentHpas {
		go HpaUpdate(hpa, pods)
	}

}

var previousHpas []apiobject.HpaStore

func WatchHpa() {
	interval := 10 * time.Second

	for {
		startTime := time.Now() // Record when HpaRoutine starts

		HpaRoutine() // Execute the routine

		elapsed := time.Since(startTime) // Calculate how long the routine took
		if elapsed < interval {
			time.Sleep(interval - elapsed) // Wait for the remainder of the interval, if any
		}
	}
}
