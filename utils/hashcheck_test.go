package utils

import "testing"

func Test(t *testing.T) {
	var tests = []struct {
		path, filename string
		want           bool
	}{
		{"/path/to/files/", "test", true},
		{"/path/to/files/", "testing", false},
	}
	for _, c := range tests {
		got := CheckForHash(c.path, c.filename)
		if got != c.want {
			t.Errorf("CheckForHash(%q, %q) == %t, want %t", c.path, c.filename, got, c.want)
		}
	}
}
