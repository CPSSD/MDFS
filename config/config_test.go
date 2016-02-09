package config

import "testing"

func Test(t *testing.T) {
	var tests = []struct {
		filename, protocol, port, path string
	}{
		{"tcp-server-conf.json", "tcp", ":8081", "/path/to/files/"},
	}

	for _, c := range tests {
		got := ParseConfiguration(c.filename)
		if got.Path != c.path || got.Port != c.port || got.Protocol != c.protocol {
			t.Error("Configuration variables does not contain expected values")
		}
	}
}
