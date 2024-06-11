package configs

const (
	LOCAL_API_SERVER_IP  = "127.0.0.1"
	MASTER_API_SERVER_IP = "192.168.1.18" // vahag_node
	API_SERVER_PORT      = "8081"
	SCHEMA               = "http://"
	LOCAL_API_URL        = SCHEMA + LOCAL_API_SERVER_IP + ":" + API_SERVER_PORT
	MASTER_API_URL       = SCHEMA + MASTER_API_SERVER_IP + ":" + API_SERVER_PORT
	masterNode           = true
	API_VERSION          = "v1"
)

func GetApiServerUrl() string {
	if masterNode {
		return LOCAL_API_URL
	} else {
		return MASTER_API_URL
	}
}

const (
	NodeUrl  = "/node"
	NodesUrl = "/nodes"

	PodUrl           = "/pod"
	PodStoreUrl      = "/podStore"
	PodStoreStatsUrl = "/podStoreStats"
	PodsURL          = "/pods"

	ServiceUrl      = "/service"
	ServicesUrl     = "/services"
	ServiceStoreURL = "/serviceStore"

	DeploymentUrl  = "/deployment"
	DeploymentsUrl = "/deployments"

	EndpointsURL = "/endpoints"

	HpaUrl      = "/hpa"
	HpaStoreUrl = "/hpaStore"
	HpasUrl     = "/hpas"

	DnsUrl  = "/dns"
	DnssUrl = "/dnss"
)
