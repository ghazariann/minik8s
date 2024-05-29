package handlers

import (
	"encoding/json"
	"fmt"
	"minik8s/internal/apiobject"
	"minik8s/internal/apiserver/etcdclient"
	"minik8s/internal/configs"
	"net/http"

	"github.com/google/uuid"
)

func AddNode(w http.ResponseWriter, r *http.Request) {
	var node apiobject.Node
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, _ := etcdclient.KeyExists(configs.ETCDNodePath + node.Metadata.Name)
	if res {
		http.Error(w, "node already exists", http.StatusConflict)
		return
	}
	node.Metadata.UUID = uuid.New().String()

	nodeStore := node.ToNodeStore()

	nodeStoreJson, err := json.Marshal(nodeStore)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO add namespace + name
	if err := etcdclient.PutKey(configs.ETCDNodePath+node.Metadata.Name, string(nodeStoreJson)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "node created: %s", node.Metadata.Name)
}
