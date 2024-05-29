package apiserver

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    //"bytes"
    //"io/ioutil"
    "go.etcd.io/etcd/client/v3"
    "time"
    "minik8s/internal/apiobject"
    "minik8s/internal/endpoints"
)

var cli *clientv3.Client

func init() {
    var err error
    cli, err = clientv3.New(clientv3.Config{
        Endpoints:   []string{"localhost:2379"},
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        log.Fatalf("Failed to connect to etcd: %v", err)
    }
}

func StartServer() {
    http.HandleFunc("/pods", handlePods)
    http.HandleFunc("/all-pods", handleAllPods) 
    http.HandleFunc("/unscheduled-pods", handleUnscheduledPods) 
    http.HandleFunc("/updatePod", handleUpdatePod)
    http.HandleFunc("/services", handleServices)
    http.HandleFunc("/all-services", handleAllServices)
    fmt.Println("API Server starting on port 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}

func handlePods(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        podName := r.URL.Query().Get("name")
        if podName == "" {
            http.Error(w, "Pod name is required", http.StatusBadRequest)
            return
        }
        podData, err := getKey("pods/" + podName)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        fmt.Fprintf(w, "Pod Data: %s", podData)
    case "POST":
        var pod apiobject.Pod
        if err := json.NewDecoder(r.Body).Decode(&pod); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        jsonData, err := json.Marshal(pod)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if err := putKey("pods/"+pod.Name, string(jsonData)); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
	 // Forward the pod information to kubelet for starting the pod
       // if err := forwardToKubelet(jsonData); err != nil {
         //   http.Error(w, "Failed to send pod start request to kubelet: "+err.Error(), http.StatusInternalServerError)
          //  return
        //}
        w.WriteHeader(http.StatusCreated)
        fmt.Fprintf(w, "Pod created: %s", pod.Name)
    default:
        http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
    }
}

func handleAllPods(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
        return
    }
    podsData, err := getAllPods()
    if err != nil {
        http.Error(w, "Failed to fetch all pods: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(podsData)
}

func handleUnscheduledPods(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
        return
    }
    podsData, err := getUnscheduledPods()
    if err != nil {
        http.Error(w, "Failed to fetch unscheduled pods: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(podsData)
}

// handleUpdatePod 处理更新 Pod 状态和 NodeName 的请求
func handleUpdatePod(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
        return
    }

    var updateRequest struct {
        Name     string `json:"name"`
        NodeName string `json:"nodeName"`
        Status   string `json:"status"`
    }
    if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := updatePod(updateRequest.Name, updateRequest.NodeName, updateRequest.Status); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Pod %s updated successfully", updateRequest.Name)
}

func handleServices(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case  "POST" :
    var service apiobject.Service
    if err := json.NewDecoder(r.Body).Decode(&service); err != nil {
        http.Error(w, "Invalid service data", http.StatusBadRequest)
        return
    }

    // 将服务数据转换为 JSON 并存储
    serviceData, err := json.Marshal(service)
    if err != nil {
        http.Error(w, "Failed to encode service data", http.StatusInternalServerError)
        return
    }
    serviceKey := "services/" + service.Name
    if err := putKey(serviceKey, string(serviceData)); err != nil {
        http.Error(w, "Failed to store service in etcd", http.StatusInternalServerError)
        return
    }

     // 获取所有Pods并筛选符合Service选择器的Pods
    podsData, err := getAllPods()
    if err != nil {
            http.Error(w, "Failed to fetch pods for service endpoints", http.StatusInternalServerError)
            return
        }
    var allPods []apiobject.Pod
    json.Unmarshal(podsData, &allPods)
    matchedPods := filterPodsBySelector(allPods, service.Selector)

    // 创建并存储对应的 Endpoint 对象
    var ep endpoints.Endpoint
    ep.ServiceName = service.Name
    for _, pod := range matchedPods {
            ep.IPs = append(ep.IPs, pod.IP) // pod结构中有IP字段
        }
    epData, err := json.Marshal(ep)
    if err != nil {
        http.Error(w, "Failed to encode endpoint data", http.StatusInternalServerError)
        return
    }
    epKey := "endpoints/" + service.Name
    if err := putKey(epKey, string(epData)); err != nil {
        http.Error(w, "Failed to store endpoint in etcd", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "Service and endpoint created successfully")

	case  "GET" :
    serviceName := r.URL.Query().Get("name")
    if serviceName == "" {
            http.Error(w, "Service name is required", http.StatusBadRequest)
            return
        }
        serviceData, err := getKey("services/" + serviceName)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "Service Data: %s", serviceData)

    }
}

func handleAllServices(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
        return
    }
    servicesData, err := getAllKeys("services/")
    if err != nil {
        http.Error(w, "Failed to fetch services: "+err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write(servicesData)
}


// updatePod 更新存储在 etcd 中的 Pod 状态和 NodeName
func updatePod(name, nodeName, status string) error {
    podKey := "pods/" + name
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    resp, err := cli.Get(ctx, podKey)
    if err != nil {
        return err
    }
    if len(resp.Kvs) == 0 {
        return fmt.Errorf("pod not found")
    }

    var pod apiobject.Pod
    if err := json.Unmarshal(resp.Kvs[0].Value, &pod); err != nil {
        return err
    }

    pod.NodeName = nodeName
    pod.Status = status

    jsonData, err := json.Marshal(pod)
    if err != nil {
        return err
    }

    _, err = cli.Put(ctx, podKey, string(jsonData))
    return err
}

func getAllPods() ([]byte, error) {
    return getPodsByCondition("")
}

func getUnscheduledPods() ([]byte, error) {
    return getPodsByCondition("NodeName")
}

func getPodsByCondition(filterKey string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    resp, err := cli.Get(ctx, "pods/", clientv3.WithPrefix())
    if err != nil {
        return nil, err
    }
    var podse []apiobject.Pod
    for _, kv := range resp.Kvs {
        var pod apiobject.Pod
        if err := json.Unmarshal(kv.Value, &pod); err == nil {
            if filterKey == "" || pod.NodeName =="" {
                podse = append(podse, pod)
            }
        }
    }
    return json.Marshal(podse)
}


func getKey(key string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    resp, err := cli.Get(ctx, key)
    if err != nil {
        return "", err
    }
    if len(resp.Kvs) > 0 {
        return string(resp.Kvs[0].Value), nil
    }
    return "", nil
}

func putKey(key, value string) error {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    _, err := cli.Put(ctx, key, value)
    return err
}

func getAllKeys(prefix string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    resp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
    if err != nil {
        return nil, err
    }
    var results []json.RawMessage
    for _, kv := range resp.Kvs {
        results = append(results, kv.Value)
    }
    return json.Marshal(results)
}

func filterPodsBySelector(pods []apiobject.Pod, selector map[string]string) (selectedPods []apiobject.Pod) {
    for _, pod := range pods {
        matches := true
        for key, value := range selector {
            if pod.Labels[key] != value {
                matches = false
                break
            }
        }
        if matches {
            selectedPods = append(selectedPods, pod)
        }
    }
    return selectedPods
}
//func forwardToKubelet(data []byte) error {
  //  resp, err := http.Post("http://localhost:10250/startPod", "application/json", bytes.NewBuffer(data))
    //if err != nil {
      //  return err
    //}
    //defer resp.Body.Close()
    //if resp.StatusCode != http.StatusOK {
      //  body, _ := ioutil.ReadAll(resp.Body)
        //return fmt.Errorf("kubelet returned error: %s", body)
    //}
    //return nil
//}
