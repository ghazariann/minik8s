
./kubectl create deployment -f testdata/deployment.yaml
./kubectl create service -f testdata/service.yaml
./kubectl create pod -f testdata/podhpa.yaml
./kubectl create hpa -f testdata/hpa.yaml

echo "Waiting..."
sleep 10 
echo docker ps
docker ps
./kubectl get pods
./kubectl get services
# ./kubectl get endpoints
./kubectl get deployments
./kubectl get hpas
