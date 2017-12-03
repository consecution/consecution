package chain

import (
	"testing"
)

func TestFilterFile(t *testing.T) {
	_, err := ChainFile("chain.yaml")
	if err != nil {
		t.Error(err)
	}
}
