package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"path"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetEndpoints(key, value string) ([]apiobject.Endpoint, error) {
	endpointsKeyValueUrl := path.Join(configs.ETCDEndpointPath, key, value)

	resp, err := etcdclient.Cli.Get(context.Background(), endpointsKeyValueUrl, clientv3.WithPrefix())

	if err != nil {
		log.Println("Failed to fetch endpoints: " + err.Error())
		return nil, err
	}

	var endpoints []apiobject.Endpoint

	for _, kv := range resp.Kvs {
		var endpoint apiobject.Endpoint
		if err := json.Unmarshal(kv.Value, &endpoint); err != nil {
			return nil, err
		}
		// if endpoint.IP == "10.40.0.3\n" || endpoint.IP == "10.40.0.1\n" || endpoint.IP == "10.40.0.0\n" {
		// 	etcdclient.DeleteKey(path.Join(endpointsKeyValueUrl, endpoint.PodUUID))
		// }
		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

func UpdateEndPoints(pod *apiobject.PodStore) error {
	for key, value := range pod.Metadata.Labels {
		UpdateEndPointsByLabel(pod, key, value)
	}
	return nil
}
func UpdateEndPointsByLabel(pod *apiobject.PodStore, key, value string) {
	endpointsKeyValueUrl := path.Join(configs.ETCDEndpointPath, key, value)

	// Fetch existing endpoints for the label
	allEndpoints, _ := GetEndpoints(key, value)

	if !podEndpointExists(allEndpoints, pod.Metadata.UUID) {
		newEndpoint, _ := createEndpoint(pod)
		endpointJson, _ := json.Marshal(newEndpoint)
		etcdclient.PutKey(path.Join(endpointsKeyValueUrl, newEndpoint.PodUUID), string(endpointJson))
		allEndpoints = append(allEndpoints, newEndpoint)
	}

	// Update service endpoints
	updateServiceEndpoints(key, value, allEndpoints)
}

// Check if the pod's endpoint already exists in the list of endpoints
func podEndpointExists(endpoints []apiobject.Endpoint, podUUID string) bool {
	for _, endpoint := range endpoints {
		if endpoint.PodUUID == podUUID {
			return true
		}
	}
	return false
}

// Create a new endpoint based on the given pod
func createEndpoint(pod *apiobject.PodStore) (apiobject.Endpoint, error) {
	myuuid, _ := uuid.NewUUID()
	endpoint := apiobject.Endpoint{
		APIObject: apiobject.APIObject{
			Metadata: apiobject.Metadata{
				UUID: myuuid.String(),
			},
		},
		IP:      pod.Status.PodIP,
		Ports:   extractPorts(pod),
		PodUUID: pod.Metadata.UUID,
	}
	return endpoint, nil
}

// Extract ports from the pod's containers
func extractPorts(pod *apiobject.PodStore) []string {
	var ports []string
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			ports = append(ports, fmt.Sprint(port.ContainerPort))
		}
	}
	return ports
}

// Update service endpoints
func updateServiceEndpoints(key, value string, allEndpoints []apiobject.Endpoint) error {
	serviceLRs, err := etcdclient.Cli.Get(context.Background(), configs.ETCDServiceSelectorPath, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, serviceLR := range serviceLRs.Kvs {
		var serviceStore apiobject.ServiceStore
		if err := json.Unmarshal(serviceLR.Value, &serviceStore); err != nil {
			return err
		}
		if serviceStore.Spec.Selector[key] != value {
			continue
		}
		serviceStore.Status.Endpoints = allEndpoints

		serviceJson, err := json.Marshal(serviceStore)
		if err != nil {
			return err
		}

		svcSelectorURL := path.Join(configs.ETCDServiceSelectorPath, key, value, serviceStore.Metadata.UUID)
		if err := etcdclient.PutKey(svcSelectorURL, string(serviceJson)); err != nil {
			return err
		}

		// Update service store (override)
		svcURL := path.Join(configs.ETCDServicePath, serviceStore.Metadata.Name)
		if err := etcdclient.PutKey(svcURL, string(serviceJson)); err != nil {
			return err
		}
	}

	return nil
}
