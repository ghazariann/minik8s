package configs

const (
	ETCDAlocIPPath = "/minik8s/allocatedClusterIP"
	///minik8s/pods/<pod-name>
	ETCDPodPath = "/minik8s/pods/"
	///minik8s/services/<service-name>
	ETCDServicePath = "/minik8s/services/"
	//minik8s/service_labels/<label-key>/<label-value>/<service-uuid>
	ETCDServiceSelectorPath = "/minik8s/service_labels/"
	//minik8s/endpoints/<label-key>/<label-value>/<pod-uuid>
	ETCDEndpointPath = "/minik8s/endpoints/"
	//minik8s/deployments/<deployment-name>
	ETCDDeploymentPath = "/minik8s/deployments/"
	//minik8s/nodes/<node-name>
	ETCDNodePath = "/minik8s/nodes/"
	//minik8s/hpas/<hpa-name>
	ETCDHpaPath = "/minik8s/hpas/"
	//minik8s/dns/<dns-name>
	ETCDDnsPath      = "/minik8s/dns/"
	ETCDDnsServiceIP = "minik8s/dnsServiceIp"
)
