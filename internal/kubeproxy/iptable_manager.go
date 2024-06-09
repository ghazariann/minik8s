package kubeproxy

import (
	"fmt"
	"log"
	"math/rand"
	"minik8s/internal/apiobject"
	"strconv"
	"time"

	"github.com/coreos/go-iptables/iptables"
)

func GenerateRandomStr(length int) string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rng.Intn(len(letterRunes))]
	}
	return string(b)
}

type IptableManager struct {
	ipt            *iptables.IPTables
	stragegy       string
	serviceToPod   map[string][]string
	serviceToChain map[string][]string
	chainToRule    map[string][]string
}

func (im *IptableManager) Initialize_iptables() {
	im.ipt, _ = iptables.New()

	im.ipt.ChangePolicy("nat", "PREROUTING", "ACCEPT")
	im.ipt.ChangePolicy("nat", "INPUT", "ACCEPT")
	im.ipt.ChangePolicy("nat", "OUTPUT", "ACCEPT")
	im.ipt.ChangePolicy("nat", "POSTROUTING", "ACCEPT")

	im.ipt.NewChain("nat", "MINIK8S-SERVICES")
	im.ipt.NewChain("nat", "MINIK8S-POSTROUTING")
	im.ipt.NewChain("nat", "MINIK8S-MARK-MASQ")

	im.ipt.Append("nat", "PREROUTING", "-j", "MINIK8S-SERVICES", "-m", "comment", "--comment", "miniK8s service portals")
	im.ipt.Append("nat", "OUTPUT", "-j", "MINIK8S-SERVICES", "-m", "comment", "--comment", "miniK8s service portals")
	im.ipt.Append("nat", "POSTROUTING", "-j", "MINIK8S-POSTROUTING", "-m", "comment", "--comment", "miniK8s postrouting rules")

	im.ipt.AppendUnique("nat", "MINIK8S-MARK-MASQ", "-j", "MARK", "--or-mark", "0x4000")
	im.ipt.AppendUnique("nat", "MINIK8S-POSTROUTING", "-m", "comment", "--comment", "miniK8s service traffic requiring SNAT", "-j", "MASQUERADE", "-m", "mark", "--mark", "0x4000/0x4000")
	log.Printf("KUBEPROXY init iptables success")
}

func (im *IptableManager) CreateService(service apiobject.ServiceStore) error {
	var clusterIp = service.Spec.ClusterIP
	seviceName := service.Metadata.Name
	ports := service.Spec.Ports
	var pod_ip_list []string
	for _, endpoint := range service.Status.Endpoints {
		pod_ip_list = append(pod_ip_list, endpoint.IP)
	}

	for _, eachports := range ports {
		log.Printf("KUBEPROXY: port: " + strconv.Itoa(eachports.Port))
		port := eachports.Port
		protocol := eachports.Protocol
		targetPort := eachports.TargetPort
		err := im.setIPTablesClusterIp(seviceName, clusterIp, port, protocol, targetPort, pod_ip_list, eachports.NodePort)
		if err != nil {
			log.Printf("KUBEPROXY: CreateService: setIPTablesClusterIp failed")
			return err
		}
	}

	for _, endpoint := range service.Status.Endpoints {
		log.Printf("KUBEPROXY: serviceToPod: " + seviceName + " " + endpoint.PodUUID)
		im.serviceToPod[seviceName] = append(im.serviceToPod[seviceName], endpoint.PodUUID)
	}
	return nil
}

func (im *IptableManager) setIPTablesClusterIp(serviceName string, clusterIP string, port int, protocol string, targetPort int, podIPList []string, nodePort int) error {
	log.Printf("KUBEPROXY: setIPTablesClusterIp: " + serviceName + " " + clusterIP + " " + strconv.Itoa(port) + " " + protocol + " " + strconv.Itoa(targetPort))

	if im.ipt == nil {
		log.Printf("KUBEPROXY: im.iptables is nil")
		return fmt.Errorf("im.iptables is nil")
	}

	dnatRule := []string{"-p", protocol, "--dport", strconv.Itoa(nodePort), "-j", "DNAT", "--to-destination", clusterIP + ":" + strconv.Itoa(port)}
	if err := im.ipt.Append("nat", "PREROUTING", dnatRule...); err != nil {
		log.Printf("Failed to insert DNAT rule for NodePort: %s", err)
		return err
	}

	kube_service := "MINIK8S-SVC-" + GenerateRandomStr(6)
	if err := im.ipt.NewChain("nat", kube_service); err != nil {
		log.Printf("KUBEPROXY: Failed to create kube_service chain: " + err.Error())
	}
	im.serviceToChain[serviceName] = append(im.serviceToChain[serviceName], kube_service)

	if err := im.ipt.Insert("nat", "MINIK8S-SERVICES", 1, "-m", "comment", "--comment",
		serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
		"--destination", clusterIP+"/"+strconv.Itoa(16), "-j", kube_service); err != nil {
		log.Printf("KUBEPROXY: Failed to insert MINIK8S-SERVICES rule for kube_service chain: " + err.Error())
		// return err
	}
	im.chainToRule[kube_service] = append(im.chainToRule[kube_service], "MINIK8S-SERVICES")

	if err := im.ipt.Insert("nat", "MINIK8S-SERVICES", 1, "-m", "comment", "--comment",
		serviceName+": cluster IP", "-p", protocol, "--dport", strconv.Itoa(port),
		"-j", "MINIK8S-MARK-MASQ", "--destination", clusterIP+"/"+strconv.Itoa(16)); err != nil {
		log.Printf("KUBEPROXY: Failed to insert MINIK8S-SERVICES rule for MINIK8S-MARK-MASQ chain: " + err.Error())
		// return err
	}
	im.chainToRule["MINIK8S-MARK-MASQ"] = append(im.chainToRule["MINIK8S-MARK-MASQ"], "MINIK8S-SERVICES")

	podNum := len(podIPList)
	log.Printf("KUBEPROXY: podNum is " + strconv.Itoa(podNum))
	for i := podNum - 1; i >= 0; i-- {
		kube_endpoint := "MINIK8S-SEP-" + GenerateRandomStr(6)
		if err := im.ipt.NewChain("nat", kube_endpoint); err != nil {
			log.Printf("KUBEPROXY: Failed to create kube_endpoint chain: " + err.Error())
		}
		im.serviceToChain[serviceName] = append(im.serviceToChain[serviceName], kube_endpoint)

		if im.stragegy == "random" {
			var prob float64 = 1 / (float64)(podNum-i)
			if i == podNum-1 { // first one
				if err := im.ipt.Insert("nat", kube_service, 1, "-j", kube_endpoint); err != nil {
					log.Printf("KUBEPROXY: Failed to create kube_service chain: " + err.Error())
					return err
				}
			} else {
				if err := im.ipt.Insert("nat", kube_service, 1, "-j", kube_endpoint,
					"-m", "statistic", "--mode", "random", "--probability", strconv.FormatFloat(prob, 'f', -1, 64)); err != nil {
					log.Printf("KUBEPROXY: Failed to create kube_service chain: " + err.Error())
					return err
				}
			}
			im.chainToRule[kube_service] = append(im.chainToRule[kube_service], kube_endpoint)
		} else if im.stragegy == "roundrobin" {
			if i == podNum-1 {
				if err := im.ipt.Insert("nat", kube_service, 1, "-j", kube_endpoint); err != nil {
					log.Printf("KUBEPROXY: Failed to create kube_service chain: " + err.Error())
					return err
				}
			} else {
				if err := im.ipt.Insert("nat", kube_service, 1, "-j", kube_endpoint,
					"-m", "statistic", "--mode", "nth", "--every", strconv.Itoa(podNum-i)); err != nil {
					log.Printf("KUBEPROXY: Failed to create kube_service chain: " + err.Error())
					return err
				}
			}
			im.chainToRule[kube_service] = append(im.chainToRule[kube_service], kube_endpoint)
		}

		if err := im.ipt.Insert("nat", kube_endpoint, 1, "-j", "DNAT",
			"-p", protocol,
			"--to-destination", podIPList[i]+":"+strconv.Itoa(targetPort)); err != nil {
			log.Printf("KUBEPROXY: Failed to create kube_service chain: " + err.Error())
			return err
		}
		im.chainToRule[kube_endpoint] = append(im.chainToRule[kube_endpoint], "DNAT")

		if err := im.ipt.Insert("nat", kube_endpoint, 1, "-j", "MINIK8S-MARK-MASQ",
			"-s", podIPList[i]+"/"+strconv.Itoa(16)); err != nil {
			log.Printf("KUBEPROXY: Failed to create kube_service chain: " + err.Error())
			return err
		}

	}
	log.Printf("KUBEPROXY: iptables rules have been set for service: " + serviceName)

	return nil
}

func (im *IptableManager) CleanIpTables(serviceName string) error {

	chainList := im.serviceToChain[serviceName]
	for _, chain := range chainList {
		im.ipt.ClearChain("nat", chain)
		im.ipt.DeleteChain("nat", chain)
	}
	im.serviceToChain[serviceName] = nil
	im.serviceToPod[serviceName] = nil
	return nil
}
