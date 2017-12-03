package runner

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/consecution/consecution/chain"
	"github.com/consecution/consecution/etcd"
	"github.com/consecution/consecution/nats"
)

var (
	DefaultQue         = 2
	ContainerDirectory = ""
)

func init() {
	home := os.Getenv("HOME")
	ContainerDirectory = fmt.Sprintf("%v/native_conatiners", home)
}

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
		err := GetImage(l)
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
	que    []*Command
	lock   *sync.Mutex
	buffer *int
}

func NewWorker(l chain.Link, i *int) Worker {
	w := Worker{
		buffer: i,
		lock:   &sync.Mutex{},
		que:    make([]*Command, 0),
		link:   l,
	}
	return w
}

func (w *Worker) UpdateQue() error {
	i := len(w.que)
	for i < 2 {
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
	fmt.Println("popping command")
	c := w.que[0]
	w.que = w.que[1:]
	w.lock.Unlock()
	fmt.Println("popped, running")
	start := time.Now()
	_, err := io.Copy(c.in, in)
	if err != nil {
		log.Println(err)
	}
	c.in.Close()
	_, err = io.Copy(out, c.out)
	if err != nil {
		log.Println(err)
	}
	_, err = io.Copy(er, c.err)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("running command")
	err = c.cmd.Wait()
	fmt.Printf("request finished in %v\n", time.Since(start))
	go func() {
		err := w.UpdateQue()
		if err != nil {
			log.Println(err)
		}
	}()
	return err
}

type Command struct {
	cmd *exec.Cmd
	in  io.WriteCloser
	out io.ReadCloser
	err io.ReadCloser
}

func NewCommand(l chain.Link) (*Command, error) {
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
	cmdline := []string{
		"run",
		"--rm",
		"-i",
		l.Image,
		l.Command,
	}
	cmdline = append(cmdline, l.Arguments...)
	c.cmd = exec.Command("docker", cmdline...)
	c.in, err = c.cmd.StdinPipe()
	if err != nil {
		return &c, err
	}
	c.out, err = c.cmd.StdoutPipe()
	if err != nil {
		return &c, err
	}
	c.err, err = c.cmd.StderrPipe()
	if err != nil {
		return &c, err
	}
	err = c.cmd.Start()
	return &c, err

}

func GetImage(l chain.Link) error {
	targetdir := fmt.Sprintf("%v/%v", ContainerDirectory, l.Image)
	_, err := os.Stat(targetdir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		fmt.Printf("Already done for  %v\n", l.Image)
		return err
	}
	fmt.Printf("pulling %v\n", l.Image)
	cmd := exec.Command("docker", "pull", l.Image)
	err = cmd.Run()
	if err != nil {
		return err
	}
	fmt.Printf("creating container %v\n", l.Image)
	cmd = exec.Command("docker", "create", "--name", l.Image, l.Image)
	if err != nil {
		return err
	}
	fmt.Printf("exporting container %v\n", l.Image)
	tarfile := fmt.Sprintf("/tmp/%v.tar", l.Image)
	cmd = exec.Command("docker", "export", "-o", tarfile, l.Image)
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = os.MkdirAll(targetdir, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Printf("untarring container %v\n", l.Image)
	cmd = exec.Command("tar", "xf", tarfile, "-C", targetdir)
	err = cmd.Run()
	return err
}
