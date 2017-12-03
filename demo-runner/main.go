package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/consecution/consecution/runner"
)

func main() {
	etcd := []string{"http://localhost:2379"}
	nats := "nats://localhost:4222"
	_, err := runner.New(nats, etcd)
	if err != nil {
		log.Fatal(err)
	}
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println(sig)
		done <- true
	}()
	<-done
}
