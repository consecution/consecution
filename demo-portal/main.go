package main

import (
	"log"

	"github.com/consecution/consecution/portal"
)

func main() {
	etcd := []string{"http://localhost:2379"}
	nats := "nats://localhost:4222"
	file := "chain.yaml"
	_, err := portal.New(file, nats, etcd)
	log.Print(err)
}
