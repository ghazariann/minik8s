package kubelet

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os/exec"
    "minik8s/internal/apiobject"
)

type Kubelet struct {
    NodeName string
}

func NewKubelet(nodeName string) *Kubelet {
    return &Kubelet{NodeName: nodeName}
}

func (k *Kubelet) StartServer() {
    http.HandleFunc("/startPod", k.handleStartPod)
    log.Println("Kubelet is listening on port 10250...")
    log.Fatal(http.ListenAndServe(":10250", nil))
}

func (k *Kubelet) handleStartPod(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
        return
    }
    var pod apiobject.Pod
    if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if err := k.StartPod(&pod); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "Pod started successfully")
}

func (k *Kubelet) StartPod(pod *apiobject.Pod) error {
    for _, container := range pod.Containers {
        // 创建一个包含所有启动参数的切片
        args := append([]string{"run", "-d"}, container.Image)
        args = append(args, container.Command...) // 将容器的命令参数追加到启动命令中
        cmd := exec.Command("docker", args...)
        if err := cmd.Start(); err != nil {
            fmt.Printf("Failed to start container %s: %v\n", container.Image, err)
            continue // 不返回错误，而是继续尝试启动其他容器
        }
    }
    return nil
}


