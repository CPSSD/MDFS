package config_test

import (
	"fmt"
	"github.com/CPSSD/MDFS/config"
	"github.com/CPSSD/MDFS/utils"
	"testing"
)

func TestConfig(t *testing.T) {
	var tests = []struct {
		filename, protocol, host, port, path string
	}{
		{utils.GetUserHome() + "/.mdfs/stnode/.stnode_conf.json", "tcp", "localhost", "8081", utils.GetUserHome() + "/.mdfs/stnode"},
		{utils.GetUserHome() + "/.mdfs/mdservice/.mdservice_conf.json", "tcp", "localhost", "1994", utils.GetUserHome() + "/.mdfs/mdservice"},
	}
	for _, c := range tests {
		got := config.ParseConfiguration(c.filename)
		if got.Path != c.path || got.Host != c.host || got.Port != c.port || got.Protocol != c.protocol {
			t.Error("Configuration variables does not contain expected values\n")
			fmt.Printf("%s,%s,%s,%s,%s,%s,%s,%s\n", got.Path, c.path, got.Host, c.host, got.Port, c.port, got.Protocol, c.protocol)
		}
	}
}
