package main

import (
	"minik8s/internal/apiserver"
	"minik8s/internal/controller"
	"minik8s/internal/kubelet"
	"minik8s/internal/kubeproxy"
)

func main() {

	kubeletInstance, _ := kubelet.NewKubelet()
	go kubeletInstance.StartServer()
	kubeproxyInstance, _ := kubeproxy.NewKubeProxy()
	go kubeproxyInstance.WatchService()

	go controller.WatchDeployment()
	apiserver.StartServer()
}
