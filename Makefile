build:
	go build -o ./kubectl ./cmd/kubectl

run:
	etcd &
	go run cmd/apiserver/main.go

.PHONY: build
