package utils

import (
	"testing"
)

// needs to be fixed
func TestKeyGen(t *testing.T) {
	var tests = []struct {
		equal bool
	}{
		{true},
	}
	for _, c := range tests {
		got, err := GenUserKeys()
		if got != c.equal {
			t.Error(err)
		}
	}
}
