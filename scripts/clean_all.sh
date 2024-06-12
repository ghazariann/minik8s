#!/bin/bash

# Execute clean_apiobjects.sh
./scripts/clean_apiobjects.sh

# Execute clean_iptables.sh
./scripts/clean_iptables.sh

# Execute clean_etcd.sh
./scripts/clean_etcd.sh

echo "{}" > "/root/minik8s/persist/known_pods.json"
echo "{}" > "/root/minik8s/persist/known_containers.json"
echo "Docker containers: (sleeping 10 seconds)"
sleep 10
docker ps

./kubectl create node -f testdata/node2.yaml
