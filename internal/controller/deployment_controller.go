package controller

import (
	"fmt"
	"io"
	"log"
	"minik8s/internal/apiobject"
	"net/http"
)

type DeploymentController interface {
	Run()
}

type deploymentController struct {
}

func (rc *deploymentController) GetAllDeploymentsFromAPIServer() ([]apiobject.DeploymentStore, error) {
	url := "http://localhost:8080/deployments"

	allDeployments := make([]apiobject.DeploymentStore, 0)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list deployments: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))

	return allDeployments, nil
}
