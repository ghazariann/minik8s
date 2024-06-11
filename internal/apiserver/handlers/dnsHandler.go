package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"net/http"
	"path"

	"github.com/google/uuid"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GetDnss(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}
	// etcdclient.Cli.Delete(context.Background(), configs.ETCDDnsPath, clientv3.WithPrefix())
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDDnsPath, clientv3.WithPrefix())
	if err != nil {
		http.Error(w, "Failed to fetch deployments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Initialize a slice to hold the decoded deployment objects
	var DnsStores []apiobject.DnsStore

	// Iterate through each key-value pair returned from the store
	for _, kv := range resp.Kvs {
		var DnsStore apiobject.DnsStore
		if err := json.Unmarshal(kv.Value, &DnsStore); err != nil {
			http.Error(w, "Error decoding deployment data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		DnsStores = append(DnsStores, DnsStore)
	}

	// Convert the deployments slice to JSON
	DnsStoreJson, err := json.Marshal(DnsStores)
	if err != nil {
		http.Error(w, "Error encoding deployment data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Println("dnss fetched successfully")
	// Set content type and send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(DnsStoreJson)

}
func AddDns(w http.ResponseWriter, r *http.Request) {
	var dns apiobject.Dns
	if err := json.NewDecoder(r.Body).Decode(&dns); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, _ := etcdclient.KeyExists(configs.ETCDDnsPath + dns.Metadata.Name)
	if res {
		http.Error(w, "Dns already exists", http.StatusConflict)
		return
	}

	for i, p := range dns.Spec.Paths {
		serviceRes, _ := etcdclient.GetKey(configs.ETCDServicePath + p.ServiceName)

		if serviceRes == "" {
			http.Error(w, "Service "+p.ServiceName+" does not exists", http.StatusConflict)
			return
		}
		service := apiobject.ServiceStore{}
		_ = json.Unmarshal([]byte(serviceRes), &service)

		dns.Spec.Paths[i].ServiceIp = service.Spec.ClusterIP
	}
	dns.Metadata.UUID = uuid.New().String()

	dnsStore := dns.ToDnsStore()
	dnsStore.Status.Phase = "pending"
	dnsStoreJson, err := json.Marshal(dnsStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO add namespace + name
	if err := etcdclient.PutKey(configs.ETCDDnsPath+dns.Metadata.Name, string(dnsStoreJson)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Dns created: %s", dns.Metadata.Name)
}
func GetDns(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is GET
	if r.Method != "GET" {
		http.Error(w, "Only GET method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract dns name from the query parameters
	dnsName := r.URL.Query().Get("name")
	if dnsName == "" {
		http.Error(w, "Dns name is required", http.StatusBadRequest)
		return
	}

	// Retrieve dns data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDDnsPath+dnsName)
	if err != nil {
		http.Error(w, "Failed to fetch dns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the dns was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Dns not found", http.StatusNotFound)
		return
	}

	// Unmarshal the dns data
	var dnsStore apiobject.DnsStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &dnsStore); err != nil {
		http.Error(w, "Error decoding dns data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal the dns data to JSON
	dnsStoreJson, err := json.Marshal(dnsStore)
	if err != nil {
		http.Error(w, "Error encoding dns data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set content type and send the response

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dnsStoreJson)
	// fmt.Fprintf(w, "Dns fetched: %s", dnsName)
}

func UpdateDnsServiceIP(w http.ResponseWriter, r *http.Request) {
	var serviceIp string
	if err := json.NewDecoder(r.Body).Decode(&serviceIp); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err := etcdclient.Cli.Put(context.Background(), configs.ETCDDnsServiceIP, serviceIp)
	if err != nil {
		http.Error(w, "Failed to update dns service IP: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetDnsServiceIP(w http.ResponseWriter, r *http.Request) {
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDDnsServiceIP)
	if err != nil {
		http.Error(w, "Failed to retrieve dns service IP: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(resp.Kvs) > 0 {
		w.Write(resp.Kvs[0].Value)
	} else {
		http.Error(w, "No dns service IP found", http.StatusNotFound)
	}
}
func UpdateDnsStatus(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is PUT

	// Extract dns name from the query parameters
	dnsName := r.URL.Query().Get("name")
	if dnsName == "" {
		http.Error(w, "Dns name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Dns object
	var dns apiobject.DnsStore
	if err := json.NewDecoder(r.Body).Decode(&dns); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing dns data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDDnsPath+dnsName)
	if err != nil {
		http.Error(w, "Failed to fetch dns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the dns was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Dns not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing dns data
	var dnsStore apiobject.DnsStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &dnsStore); err != nil {
		http.Error(w, "Error decoding dns data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the dns data ( status in running and has weave IP)
	dnsStore.Status = dns.Status
	// Marshal the updated dns data
	dnsStoreJson, err := json.Marshal(dnsStore)
	if err != nil {
		http.Error(w, "Error encoding dns data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the dns in etcd
	if err := etcdclient.PutKey(configs.ETCDDnsPath+dnsName, string(dnsStoreJson)); err != nil {
		http.Error(w, "Failed to update dns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "DnsStore updated: %s", dnsName)
}

func UpdateDns(w http.ResponseWriter, r *http.Request) {

	// Extract dns name from the query parameters
	dnsName := r.URL.Query().Get("name")
	if dnsName == "" {
		http.Error(w, "Dns name is required", http.StatusBadRequest)
		return
	}

	// Decode the request body into a Dns object
	var dns apiobject.DnsStore
	if err := json.NewDecoder(r.Body).Decode(&dns); err != nil {
		http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve the existing dns data from etcd
	resp, err := etcdclient.Cli.Get(context.Background(), configs.ETCDDnsPath+dnsName)
	if err != nil {
		http.Error(w, "Failed to fetch dns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the dns was found
	if len(resp.Kvs) == 0 {
		http.Error(w, "Dns not found", http.StatusNotFound)
		return
	}

	// Unmarshal the existing dns data
	var dnsStore apiobject.DnsStore
	if err := json.Unmarshal(resp.Kvs[0].Value, &dnsStore); err != nil {
		http.Error(w, "Error decoding dns data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the dns data
	dnsStore.Spec = dns.Spec
	dnsStore.Metadata.Labels = dns.Metadata.Labels
	dnsStore.Status = dns.Status
	// Marshal the updated dns data
	dnsStoreJson, err := json.Marshal(dnsStore)
	if err != nil {
		http.Error(w, "Error encoding dns data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the dns in etcd
	if err := etcdclient.PutKey(configs.ETCDDnsPath+dnsName, string(dnsStoreJson)); err != nil {
		http.Error(w, "Failed to update dns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with confirmation
	fmt.Fprintf(w, "Dns updated: %s", dnsName)
}

func DeleteDns(w http.ResponseWriter, r *http.Request) {
	// Ensure the method is DELETE
	if r.Method != "DELETE" {
		http.Error(w, "Only DELETE method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Extract dns name from the query parameters
	dnsName := r.URL.Query().Get("name")
	if dnsName == "" {
		http.Error(w, "Dns name is required", http.StatusBadRequest)
		return
	}

	// Delete the dns from etcd
	dnsRes, err := etcdclient.GetKey(configs.ETCDDnsPath + dnsName)

	if dnsRes == "" {
		http.Error(w, "Dns "+dnsName+" does not exists: "+err.Error(), http.StatusInternalServerError)
	}
	dns := apiobject.DnsStore{}
	err = json.Unmarshal([]byte(dnsRes), &dns)
	if err != nil {
		http.Error(w, "Failed to decode dns data: "+err.Error(), http.StatusInternalServerError)
	}
	err = etcdclient.DeleteKey(configs.ETCDDnsPath + dnsName)

	if err != nil {
		http.Error(w, "Failed to delete dns: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// delete endpoints
	for key, value := range dns.Metadata.Labels {
		endpointsKVURL := path.Join(configs.ETCDEndpointPath, key, value, dns.Metadata.UUID)
		etcdclient.DeleteKey(endpointsKVURL)
	}
	// Respond with confirmation
	fmt.Fprintf(w, "Dns deleted: %s", dnsName)
}
