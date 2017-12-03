package portal

import (
	"errors"
	"fmt"

	"github.com/consecution/consecution/chain"
	"github.com/consecution/consecution/etcd"
	"github.com/consecution/consecution/nats"
)

type Portal struct {
	nats  nats.Nats
	etcd  etcd.Etcd
	chain chain.Chain
}

func New(file string, natsurl string, etcdurls []string) (Portal, error) {
	p := Portal{}
	var err error
	p.nats, err = nats.New(natsurl)
	if err != nil {
		return p, err
	}
	p.etcd, err = etcd.New(etcdurls)
	if err != nil {
		return p, err
	}
	p.chain, err = chain.ChainFile(file)
	if err != nil {
		return p, err
	}
	err = p.etcd.SetChain(p.chain)
	return p, err
}

func (p *Portal) Send(b []byte) ([]byte, error) {
	var err error
	if len(p.chain.Links) == 0 {
		return nil, errors.New("No links in chain")
	}
	for k, l := range p.chain.Links {
		b, err = p.nats.Send(l.Id(), b)
		if err != nil {
			return b, fmt.Errorf("Error on link %v-%v: %v", k, l.Name, err)
		}
	}
	return b, err
}
