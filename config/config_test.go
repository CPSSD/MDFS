package config

import "testing"

func TestConfig(t *testing.T) {
	var tests = []struct {
		filename, protocol, host, port, path string
	}{
		{"/path/to/files/config/stnodeconf.json", "tcp", "localhost", "8081", "/path/to/files/stnode/"},
		{"/path/to/files/config/mdserviceconf.json", "tcp", "localhost", "1994", "/path/to/files/mdservice/"},


	for _, c := range tests {
		got := ParseConfiguration(c.filename)
		if got.Path != c.path || got.Host != c.host || got.Port != c.port || got.Protocol != c.protocol {
			t.Error("Configuration variables does not contain expected values")
		}
	}
}
