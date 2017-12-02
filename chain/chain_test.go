package chain

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestFilterFile(t *testing.T) {
	c, err := ChainFile("chain.yaml")
	if err != nil {
		t.Error(err)
	}
	spew.Dump(c)
}
