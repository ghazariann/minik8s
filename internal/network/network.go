package network

import (
	"fmt"
	"minik8s/internal/apiobject"
)

// Simulated simple network manager
var nextIP = 1

// AssignIP assigns a unique IP address to a pod
func AssignIP(pod *apiobject.Pod) {
	pod.IP = fmt.Sprintf("10.0.0.%d", nextIP)
	nextIP++
}
