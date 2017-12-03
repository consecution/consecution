package etcd

import (
	"context"
	"fmt"

	"github.com/consecution/consecution/chain"
	"github.com/coreos/etcd/client"
)

const (
	workers = "workers/"
	worker  = "workers/%s"
)

type Etcd struct {
	kapi client.KeysAPI
}

func New(endpoints []string) (Etcd, error) {
	cfg := client.Config{
		Endpoints: endpoints,
	}
	var e Etcd
	c, err := client.New(cfg)
	if err != nil {
		return e, err
	}
	e.kapi = client.NewKeysAPI(c)
	e.kapi.Set(context.Background(), workers, "", &client.SetOptions{Dir: true})
	return e, err
}

func (e Etcd) SetChain(c chain.Chain) error {
	for _, l := range c.Links {
		err := e.setLink(l)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e Etcd) setLink(l chain.Link) error {
	y, err := l.String()
	if err != nil {
		return err
	}
	fmt.Printf("Setting Link %v", l.Id())
	_, err = e.kapi.Create(context.Background(), fmt.Sprintf(worker, l.Id()), y)
	return err
}

func (e Etcd) GetLinks() (chain.Chain, error) {
	var c chain.Chain
	r, err := e.kapi.Get(context.Background(), workers, &client.GetOptions{Recursive: true})
	if err != nil {
		return c, err
	}
	for _, n := range r.Node.Nodes {
		l, err := chain.StringToLink(n.Value)
		if err != nil {
			return c, err
		}
		c.Links = append(c.Links, l)
	}
	return c, nil
}
