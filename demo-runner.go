package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/consecution/consecution/runner"
)

func main() {
	time.Sleep(5 * time.Second)
	etcd := []string{"http://etcd:2379"}
	nats := "nats://nats:4222"
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
