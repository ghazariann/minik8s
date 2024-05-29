package etcdclient

import (
	"context"
	"encoding/json"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var Cli *clientv3.Client

func init() {
	var err error
	Cli, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
}

func GetKey(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := Cli.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if len(resp.Kvs) > 0 {
		return string(resp.Kvs[0].Value), nil
	}
	return "", nil
}

func PutKey(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := Cli.Put(ctx, key, value)
	return err
}

// GetObjects retrieves a list of objects from etcd based on the provided URL and object type
func GetObjects(url string, objPtr interface{}) ([]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := Cli.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	var objects []interface{}
	for _, kv := range resp.Kvs {
		if err := json.Unmarshal(kv.Value, objPtr); err != nil {
			return nil, err
		}
		objects = append(objects, objPtr)
	}

	return objects, nil
}

func DeleteKey(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := Cli.Delete(ctx, key)
	return err
}

func UpdateKey(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := Cli.Put(ctx, key, value)
	return err
}

func KeyExists(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := Cli.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return len(resp.Kvs) > 0, nil
}
