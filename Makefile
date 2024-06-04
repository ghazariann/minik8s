build:
	go build -o ./kubectl ./cmd/kubectl

run:
	go run cmd/apiserver/main.go
	go run cmd/worker/main.go

runclient: 
	go run cmd/worker/main.go
runserver:
	go run cmd/apiserver/main.go
	
runetcd:
	etcd

.PHONY: build
