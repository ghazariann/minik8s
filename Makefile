build:
	go build -o ./kubectl ./cmd/kubectl

runserver:
	go run cmd/apiserver/main.go
	
runetcd:
	etcd
run:
	etcd &
	go run cmd/apiserver/main.go

.PHONY: build
