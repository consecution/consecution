package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/consecution/consecution/portal"
)

func main() {
	time.Sleep(5 * time.Second)
	etcd := []string{"http://etcd:2379"}
	nats := "nats://nats:4222"
	file := "/files/chain.yaml"
	p, err := portal.New(file, nats, etcd)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Fprint(w, err)
		}
		var resp []byte
		resp, err = p.Send(b)
		if err != nil {
			fmt.Fprint(w, err)
		}
		w.Write(resp)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
