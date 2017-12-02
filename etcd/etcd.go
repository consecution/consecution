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
	kapi client.KeysApi
}

func New(endpoints []string) (e, err) {
	cfg := client.Config{
		Endpoints: endpoints,
	}
	var e Etcd
	c, err := client.New(cfg)
	if err != nil {
		return e, err
	}
	e.kapi = client.NewKeysAPI(c)
	_, err = e.kapi.Set(context.Background(), workers, "", client.SetOptions{Dir: true})
	return e, err
}

func (e Etcd) SetChain(c chain.Chain) error {
	for l, _ := range c {
		err := e.SetChain(l)
		if err != nil {
			return err
		}
	}
}

func (e Etcd) setLink(l chain.Link) error {
	_, err = e.kapi.Create(context.Background(), fmt.Sprintf(worker, l.Id()), l.String)
}

func (e Etcd) GetLinks() (chain.Chain, error) {
	var c chain.Chain
	r, err := e.kapi.Get(context.Background(), workers, client.GetOptions{Recursive: true})
	if err != nil {
		return c, err
	}
	for n, _ := range r.Node.Nodes {
		l, err := chain.StringToLink(n.Value)
		if err != nil {
			return c, err
		}
		c = append(c, l)
	}
	return c, nil
}
