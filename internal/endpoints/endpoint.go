// 文件路径: internal/endpoints/endpoint.go
package endpoints

// Endpoint represents a network endpoint that includes the addresses of pods that match the service selector.
type Endpoint struct {
    ServiceName string   `json:"serviceName"`
    IPs         []string `json:"ips"`
}

