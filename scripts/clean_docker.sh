# Get IDs of all containers having "weave" in their names
weave_containers=$(docker ps -a --filter "name=weave" --format "{{.ID}}")

# Stop and delete all containers except those with "weave" in their names
docker ps -a --format "{{.ID}} {{.Names}}" | \
while read -r container_id container_name; do
    if ! echo "$weave_containers" | grep -q "$container_id"; then
        echo "Stopping and deleting container: $container_name"
        docker stop "$container_id" && docker rm -f "$container_id"
    fi
done
