package config

import (
	"testing"
	"github.com/CPSSD/MDFS/utils"
)

func TestConfig(t *testing.T) {
	var tests = []struct {
		filename, protocol, host, port, path string
	}{
		{utils.GetUserHome()+"/path/to/files/config/stnodeconf.json", "tcp", "localhost", "8081", "/path/to/files/stnode/"},
		{utils.GetUserHome()+"/path/to/files/config/mdserviceconf.json", "tcp", "localhost", "1994", "/path/to/files/mdservice/"},
	}
	for _, c := range tests {
		got := ParseConfiguration(c.filename)
		if got.Path != c.path || got.Host != c.host || got.Port != c.port || got.Protocol != c.protocol {
			t.Error("Configuration variables does not contain expected values")
		}
	}
}
