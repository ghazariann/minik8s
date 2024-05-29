package scheduler

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"minik8s/internal/apiobject"
// 	"minik8s/internal/configs"
// 	"net/http"
// 	"strconv"
// )

// type SchedulePolicy string

// const (
// 	RoundRobin SchedulePolicy = "RoundRobin"
// 	Random     SchedulePolicy = "Random"
// )

// type Scheduler struct {
// 	polocy        SchedulePolicy
// 	apiServerHost string
// 	apiServerPort int
// }

// func (sch *Scheduler) GetAllNodes() (nodes []apiobject.NodeStore, err error) {
// 	uri := configs.GetApiServerUrl() + configs.NodesURL
// 	var allNodes []apiobject.NodeStore
// 	code, err := netrequest.GetRequestByTarget(uri, &allNodes, "data")

// 	if err != nil {
// 		k8log.ErrorLog("Scheduler", "get all nodes failed "+err.Error())
// 		return nil, err
// 	}

// 	if code != http.StatusOK {
// 		k8log.ErrorLog("Scheduler", "get all nodes failed, code: "+fmt.Sprint(code))
// 		return nil, fmt.Errorf("get all nodes failed, code: %d", code)
// 	}

// 	return allNodes, nil
// }

// func (sch *Scheduler) RequestSchedule() {
// 	// TODO
// 	allNodes, err := sch.GetAllNodes()

// 	if err != nil {
// 		log.Printf("Scheduler", "获取所有节点失败"+err.Error())
// 	}

// 	// 调度的时候筛选存活的节点
// 	nodes := make([]apiobject.NodeStore, 0)
// 	for _, node := range allNodes {
// 		if node.Status.Condition == apiobject.Ready {
// 			nodes = append(nodes, node)
// 		}
// 	}

// 	// 反序列化pod
// 	podStore := &apiobject.PodStore{}
// 	err = json.Unmarshal([]byte(parsedMsg.Content), &podStore)
// 	if err != nil {
// 		log.Printf("Scheduler", "反序列化pod失败")
// 		return
// 	}

// 	var scheduledNode string

// 	// 如果在pod中指定了node
// 	if podStore.Spec.NodeName != "" {
// 		// 检查node是否存在
// 		for _, node := range nodes {
// 			if node.GetName() == podStore.Spec.NodeName {
// 				scheduledNode = podStore.Spec.NodeName
// 			}
// 		}
// 	}

// 	// 如果未指定node或者指定的node无效，则选择一个节点
// 	if scheduledNode == "" {
// 		scheduledNode = sch.ChooseFromNodes(nodes)
// 	}

// 	if scheduledNode == "" {
// 		log.Printf("Scheduler", "没有可用的节点")
// 		return
// 	}

// 	// 为pod添加node信息
// 	podStore.Spec.NodeName = scheduledNode

// 	// 更新Apiserver中的Pod信息
// 	URL := stringutil.Replace(config.PodSpecURL, config.URL_PARAM_NAMESPACE_PART, podStore.GetPodNamespace())
// 	URL = stringutil.Replace(URL, config.URL_PARAM_NAME_PART, podStore.GetPodName())
// 	URL = config.GetAPIServerURLPrefix() + URL

// 	code, _, err := netrequest.PutRequestByTarget(URL, podStore)
// 	if err != nil {
// 		log.Printf("Scheduler", "更新Pod信息失败"+err.Error())
// 		return
// 	}
// 	if code != http.StatusOK {
// 		log.Printf("Scheduler", "更新Pod信息失败,code: "+strconv.Itoa(code))
// 		return
// 	}

// 	podUpdate := &entity.PodUpdate{
// 		Action:    message.CREATE,
// 		PodTarget: *podStore,
// 		Node:      scheduledNode,
// 	}
// 	message.PublishUpdatePod(podUpdate)
// }
