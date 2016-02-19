package config

import "testing"

func TestConfig(t *testing.T) {
	var tests = []struct {
		filename, protocol, port, path string
	}{
		{"../storagenode/serverconf.json", "tcp", ":8081", "/path/to/files/"},
	}

	for _, c := range tests {
		got := ParseConfiguration(c.filename)
		if got.Path != c.path || got.Port != c.port || got.Protocol != c.protocol {
			t.Error("Configuration variables does not contain expected values")
		}
	}
}
