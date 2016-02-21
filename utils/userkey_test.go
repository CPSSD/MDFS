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
		got, err := IsKeys()
		if got != c.equal {
			t.Error("GenUserKeys() generation failed.")
			t.Error(err)
			t.Error("To test key generation, you may need\nto remove \".public_key_mdfs\" and \".private_key_mdfs\"\nfrom directory: \"/path/to/files/\"")

		}
	}
}

func IsKeys() (bool, error) {

	test1, err := GenUserKeys()
	return test1, err
}