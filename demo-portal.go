package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/consecution/consecution/portal"
)

func main() {
	etcd := []string{"http://localhost:2379"}
	nats := "nats://localhost:4222"
	file := "files/chain.yaml"
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
