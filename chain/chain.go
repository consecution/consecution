package chain

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

var (
	// ErrChainDoesNotExist the chain was not found in the list of chains
	ErrChainDoesNotExist = errors.New("Chain does not exist")
)

// Link is one command in a chain
type Link struct {
	// Image docker image full url
	Image string `yaml:"Image"`
	// Command to execute
	Command string `yaml:"Command"`
	// Arguments to command
	Arguments []string `yaml:"Arguments"`
}

// New returns a new filter with defaults set
func New() Link {
	return Link{
		Arguments: make([]string, 0),
	}
}

func StringToLink(s string) (Link, error) {
	l := New()
	err := yaml.Unmarshal([]byte(s), &l)
	return l, err
}

func (l Link) Id() string {
	v := fmt.Sprintf(l.Image, l.Command, strings.Join(l.Arguments, ""))
	return fmt.Sprintf("%x", md5.Sum([]byte(v)))
}

func (l *Link) String() (string, error) {
	v, err := yaml.Marshal(l)
	return string(v), err
}

// Argument appends an argument
func (f *Link) Argument(a string) {
	f.Arguments = append(f.Arguments, a)
}

// Chain of commands where each feeds into stdin of the next
type Chain struct {
	Links []Link `yaml:"Chain"`
}

// NewChain returns a new chain
func NewChain() Chain {
	return Chain{make([]Link, 0)}
}

// ChainFile Loads a chain from a file
func ChainFile(file string) (Chain, error) {
	c := NewChain()
	return c, fromFile(file, &c)
}

func fromFile(file string, c *Chain) error {
	_, err := os.Stat(file)
	if err != nil {
		return err
	}
	filebytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(filebytes, c)
	if err != nil {
		return err
	}
	return nil
}
