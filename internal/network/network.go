package network

import (
	"fmt"
	"math/rand"
	"os/exec"
)

type NetworkPlugin interface {
	SetupPodNetwork(podName, namespace string) (string, error)
	TeardownPodNetwork(podName, namespace string) error
}

type SimpleNetworkPlugin struct{}

func NewSimpleNetworkPlugin() *SimpleNetworkPlugin {
	return &SimpleNetworkPlugin{}
}

func (p *SimpleNetworkPlugin) SetupPodNetwork(podName, namespace string) (string, error) {
	podIP := allocateIP()
	podNamespace := fmt.Sprintf("ns-%s", podName)

	// Create network namespace
	if err := exec.Command("ip", "netns", "add", podNamespace).Run(); err != nil {
		return "", fmt.Errorf("failed to create network namespace: %v", err)
	}

	// Create veth pair
	vethHost := fmt.Sprintf("veth-%s", podName)
	vethPod := fmt.Sprintf("veth-%s-pod", podName)
	if err := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethPod).Run(); err != nil {
		return "", fmt.Errorf("failed to create veth pair: %v", err)
	}

	// Move vethPod to podNamespace
	if err := exec.Command("ip", "link", "set", vethPod, "netns", podNamespace).Run(); err != nil {
		return "", fmt.Errorf("failed to move veth to namespace: %v", err)
	}

	// Set up vethHost
	if err := exec.Command("ip", "link", "set", vethHost, "up").Run(); err != nil {
		return "", fmt.Errorf("failed to set vethHost up: %v", err)
	}

	// Assign IP to vethPod and bring it up
	if err := exec.Command("ip", "netns", "exec", podNamespace, "ip", "addr", "add", podIP, "dev", vethPod).Run(); err != nil {
		return "", fmt.Errorf("failed to assign IP to vethPod: %v", err)
	}
	if err := exec.Command("ip", "netns", "exec", podNamespace, "ip", "link", "set", vethPod, "up").Run(); err != nil {
		return "", fmt.Errorf("failed to set vethPod up: %v", err)
	}

	// Set up default route
	if err := exec.Command("ip", "netns", "exec", podNamespace, "ip", "route", "add", "default", "dev", vethPod).Run(); err != nil {
		return "", fmt.Errorf("failed to set default route: %v", err)
	}

	return podIP, nil
}

func (p *SimpleNetworkPlugin) TeardownPodNetwork(podName, namespace string) error {
	podNamespace := fmt.Sprintf("ns-%s", podName)
	vethHost := fmt.Sprintf("veth-%s", podName)

	// Delete veth pair
	if err := exec.Command("ip", "link", "del", vethHost).Run(); err != nil {
		return fmt.Errorf("failed to delete veth pair: %v", err)
	}

	// Delete network namespace
	if err := exec.Command("ip", "netns", "del", podNamespace).Run(); err != nil {
		return fmt.Errorf("failed to delete network namespace: %v", err)
	}

	return nil
}

// Mock function to allocate IP addresses (in a real scenario, you would use an IPAM)
func allocateIP() string {
	return fmt.Sprintf("10.244.%d.%d/24", rand.Intn(255), rand.Intn(255))

}
