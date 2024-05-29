package main

import (
	// "fmt"
	"minik8s/internal/apiserver"
	"minik8s/internal/controller"
	"minik8s/internal/kubelet"
	// "minik8s/internal/services"
	// "minik8s/internal/pods"
	// "minik8s/internal/network"
	// "minik8s/internal/autoscaler"
)

func main() {
	// 	myPod := pods.Pod{
	//         Name: "test-pod",
	//         Containers: []pods.Container{
	//             {Image: "busybox", Command: []string{"sleep", "1000"}, Ports: []int{80}},
	//         },
	//         Status: "Created",
	//         Labels: map[string]string{"app": "myapp"}, // 确保标签匹配Service的选择器
	//         NodeName: "",
	// }

	//     startPod(&myPod)
	//     fmt.Printf("Pod started: %s, Status: %s, IP: %s, Labels: %v\n", myPod.Name, myPod.Status, myPod.IP, myPod.Labels)

	//     scaler := autoscaler.AutoScaler{
	//         TargetPod: &myPod,
	//         MinReplicas: 1,
	//         MaxReplicas: 10,
	//         CurrentReplicas: 1,
	//         Utilization: 0.80,  // 假设当前CPU利用率为80%
	//     }

	//     scaler.Scale()  // 根据当前Utilization进行扩缩容

	//     myService := services.Service{
	//         Name:      "my-service",
	//         Selector:  map[string]string{"app": "myapp"},
	//         Port:      80,
	//         TargetPort: 8080,
	//         IP:        "10.0.0.100",
	//     }

	//     pod := myService.Forward() // 假设Forward方法正确处理标签和选择器的匹配
	//     if pod != nil {
	//         fmt.Printf("Request forwarded to Pod: %s, IP: %s\n", pod.Name, pod.IP)
	//     } else {
	//         fmt.Println("No available pods to forward the request.")
	//     }

	kubeletInstance, _ := kubelet.NewKubelet("testkubelet")
	go kubeletInstance.StartServer()
	go controller.WatchDeployment()
	apiserver.StartServer()
}

// func startPod(pod *pods.Pod) {
//     network.AssignIP(pod)  // Assign an IP address when the pod is started
//     pod.Status = "Running"
// }
