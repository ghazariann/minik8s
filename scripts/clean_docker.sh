# Stop and delete all running containers except Weave
weave_container=$(docker ps --filter "name=weave" --format "{{.ID}}")

# Stop and delete all containers except Weave
docker ps --format "{{.ID}} {{.Names}}" | \
while read -r container_id container_name; do
    if [ "$container_id" != "$weave_container" ]; then
        echo "Stopping and deleting container: $container_name"
        docker stop "$container_id" && docker rm -f "$container_id"
    fi
done