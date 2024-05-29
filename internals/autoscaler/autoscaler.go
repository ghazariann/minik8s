package autoscaler

import (
	"fmt"
	"minik8s/internal/apiobject"
)

type AutoScaler struct {
	TargetPod       *apiobject.Pod
	MinReplicas     int
	MaxReplicas     int
	CurrentReplicas int
	Utilization     float64 // 目标资源利用率（例如CPU利用率）
}

func (as *AutoScaler) Scale() {
	// 简化的扩缩容逻辑
	if as.Utilization > 0.75 && as.CurrentReplicas < as.MaxReplicas {
		as.CurrentReplicas++
		fmt.Printf("Scaling up: New replica count is %d\n", as.CurrentReplicas)
	} else if as.Utilization < 0.25 && as.CurrentReplicas > as.MinReplicas {
		as.CurrentReplicas--
		fmt.Printf("Scaling down: New replica count is %d\n", as.CurrentReplicas)
	}
}
