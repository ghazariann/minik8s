#!/bin/bash

# Execute clean_apiobjects.sh
./scripts/clean_apiobjects.sh

# Execute clean_iptables.sh
./scripts/clean_iptables.sh

# Execute clean_etcd.sh
./scripts/clean_etcd.sh

echo "Docker containers: (sleeping 10 seconds)"
sleep 10
docker ps

