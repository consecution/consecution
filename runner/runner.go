package runner

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"

	"github.com/consecution/consecution/chain"
	"github.com/consecution/consecution/etcd"
	"github.com/consecution/consecution/nats"
)

var (
	DefaultQue = 2
)

type Runner struct {
	nats    nats.Nats
	etcd    etcd.Etcd
	chain   chain.Chain
	workers map[string]Worker
	que     int
}

func New(natsurl string, etcdurls []string) (Runner, error) {
	r := Runner{}
	var err error
	r.nats, err = nats.New(natsurl)
	if err != nil {
		return r, err
	}
	r.etcd, err = etcd.New(etcdurls)
	if err != nil {
		return r, err
	}
	r.chain, err = r.etcd.GetChain()
	if err != nil {
		return r, err
	}
	r.que = DefaultQue
	r.workers = make(map[string]Worker)
	err = r.BuildWorkers()
	return r, err
}

func (r *Runner) BuildWorkers() error {
	for _, l := range r.chain.Links {
		w := NewWorker(l, &r.que)
		err := GetImage(l.Image)
		if err != nil {
			return err
		}
		err = w.UpdateQue()
		if err != nil {
			return err
		}
		err = r.nats.Register(l.Id(), &w)
		r.workers[l.Id()] = w
	}
	return nil
}

func (r *Runner) SetBuffer(i int) {
	r.que = i
	for _, w := range r.workers {
		err := w.UpdateQue()
		if err != nil {
			log.Println(err)
		}
	}
}

type Worker struct {
	link   chain.Link
	que    []Command
	lock   *sync.Mutex
	buffer *int
}

func NewWorker(l chain.Link, i *int) Worker {
	w := Worker{
		buffer: i,
		lock:   &sync.Mutex{},
		que:    make([]Command, 0),
		link:   l,
	}
	return w
}

func (w *Worker) UpdateQue() error {
	i := len(w.que)
	for i < *w.buffer {
		c, err := NewCommand(w.link)
		if err != nil {
			return err
		}
		w.lock.Lock()
		w.que = append(w.que, c)
		i = len(w.que)
		w.lock.Unlock()
	}
	return nil
}

func (w *Worker) Handle(in io.Reader, out io.Writer, er io.Writer) error {
	fmt.Println("received request")
	fmt.Println(w.link.Name)
	w.lock.Lock()
	c := w.que[0]
	w.que = w.que[1:]
	w.lock.Unlock()
	go func() {
		err := w.UpdateQue()
		if err != nil {
			log.Println(err)
		}
	}()
	io.Copy(c.in, in)
	io.Copy(out, c.out)
	io.Copy(er, c.err)
	err := c.cmd.Wait()
	return err
}

type Command struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
	err io.ReadCloser
}

func NewCommand(l chain.Link) (Command, error) {
	var c Command
	var err error
	/*
		constr := make([]string, 0)
		if l.Constraints.CPU != 0 {
			constr = []string{"--cpu", strconv.FormatFloat(l.Constraints.CPU, 'f', -1, 32)}
		}
		if l.Constraints.Memory != "" {
			constr = []string{"-m", l.Constraints.Memory}
		}
	*/
	c.cmd = exec.Command("docker", "run", "--rm", "-i", l.Image, l.Command, "'test'")
	c.in, err = c.cmd.StdinPipe()
	if err != nil {
		return c, err
	}
	c.out, err = c.cmd.StdoutPipe()
	if err != nil {
		return c, err
	}
	c.err, err = c.cmd.StderrPipe()
	if err != nil {
		return c, err
	}
	err = c.cmd.Start()
	return c, err

}

func GetImage(image string) error {
	cmd := exec.Command("docker", "pull", image)
	return cmd.Run()
}
