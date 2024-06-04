package main

import (
	"minik8s/internal/apiserver"
	"minik8s/internal/controller"
)

func main() {

	go controller.WatchDeployment()
	go controller.WatchHpa()
	apiserver.StartServer()
}
