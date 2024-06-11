package main

import (
	"minik8s/internal/kubelet"
	"minik8s/internal/kubeproxy"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	kubeletInstance, _ := kubelet.NewKubelet()
	kubeletInstance.LoadFromJSON()
	wg.Add(1)
	go func() {
		defer wg.Done()
		kubeletInstance.WatchPods()
	}()

	kubeproxyInstance, _ := kubeproxy.NewKubeProxy()
	wg.Add(1)
	// kubeproxyInstance.CreateNginxPod()
	go func() {
		defer wg.Done()
		kubeproxyInstance.WatchService()
	}()

	defer kubelet.UnRegisterNode()
	wg.Wait() // Wait for all goroutines to complete
}
