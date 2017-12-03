package nats

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"time"

	nats_core "github.com/nats-io/go-nats"
)

type Handler interface {
	Handle(in io.Reader, out io.Writer, err io.Writer) error
}

type Nats struct {
	nats *nats_core.Conn
}

func New(url string) (Nats, error) {
	var n Nats
	var err error
	n.nats, err = nats_core.Connect(url)
	return n, err
}

func (n *Nats) Register(key string, h Handler) {
	w := wrapper{h, n}
	n.nats.QueueSubscribe(key, key, w.MsgHandler)
}

func (n *Nats) Send(key string, msg []byte) ([]byte, error) {
	m, err := n.nats.Request(key, msg, 50*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if m.Data[0] == byte(1) {
		return nil, errors.New(string(m.Data[1:]))
	}
	return m.Data[1:], nil
}

type wrapper struct {
	h Handler
	n *Nats
}

func (w *wrapper) MsgHandler(msg *nats_core.Msg) {
	in := bytes.NewBuffer(msg.Data)
	var out bytes.Buffer
	var errs bytes.Buffer
	err := w.h.Handle(in, bufio.NewWriter(&out), bufio.NewWriter(&errs))
	if err != nil {
		errs = *bytes.NewBufferString(err.Error())
	}
	reply := append([]byte{0}, out.Bytes()...)
	if errs.Len() > 0 {
		reply = append([]byte{0}, errs.Bytes()...)
	}
	err = w.n.nats.Publish(msg.Reply, reply)
	if err != nil {
		log.Print(err)
	}
}
