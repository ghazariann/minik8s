#!/bin/bash

# List all keys starting with "/minik8s"
keys=$(etcdctl get /minik8s/ --prefix=true --keys-only)

# Check if any keys were found
if [ -z "$keys" ]; then
    echo "No keys found starting with '/minik8s'. Exiting."
    exit 0
fi

# Loop through each key and delete it
while IFS= read -r key; do
    echo "Deleting key: $key"
    etcdctl del "$key"
done <<< "$keys"
