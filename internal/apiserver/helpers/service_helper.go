package helpers

import (
	"encoding/json"
	"log"
	"math/rand"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"strconv"
	"strings"
	"time"
)

func AllocateNewClusterIP() (string, error) {
	// 1. Get the maximum IP from etcd
	curMaxIP, err := etcdclient.GetKey(configs.ETCDAlocIPPath)
	if err != nil {
		log.Println("KUBEPROXY: AllocateNewClusterIP failed, error: " + err.Error())
		return "", err
	}
	allocatedIPMap := make(map[int]bool)
	if len(curMaxIP) != 0 {
		// Convert the object to a map
		err = json.Unmarshal([]byte(curMaxIP), &allocatedIPMap)
		if err != nil {
			log.Println("KUBEPROXY", "AllocateNewClusterIP failed, error: "+err.Error())
			return "", err
		}
	}

	// Generate random numbers for the IP address
	num0 := strconv.Itoa(192)
	num1 := strconv.Itoa(168)
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	num2 := rng.Intn(256)
	num3 := rng.Intn(256)

	// Check if the IP has been allocated
	allocatedIPMap[num2*256+num3] = true
	allocatedIPJson, err := json.Marshal(allocatedIPMap)
	if err != nil {
		log.Println("KUBEPROXY", "AllocateNewClusterIP failed, error: "+err.Error())
		return "", err
	}
	// ctx, _ := context.WithTimeout(context.Background(), time.Second)
	err = etcdclient.PutKey(configs.ETCDAlocIPPath, string(allocatedIPJson)) // Convert allocatedIPJson to string
	if err != nil {
		log.Println("KUBEPROXY", "AllocateNewClusterIP failed, error: "+err.Error())
		return "", err
	}
	return strings.Join([]string{num0, num1, strconv.Itoa(num2), strconv.Itoa(num3)}, "."), nil
}
