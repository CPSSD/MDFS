package utils

import "testing"

func TestHash(t *testing.T) {
	var tests = []struct {
		path, filename string
		want           bool
	}{
		{"/path/to/files/", "test.txt", true},
		{"/path/to/files/", "test.jpg", true}
		{"/path/to/files/", "testing", false},
	}
	for _, c := range tests {
		got := CheckForHash(c.path, c.filename)
		if got != c.want {
			t.Errorf("CheckForHash(%q, %q) == %t, want %t", c.path, c.filename, got, c.want)
		}
	}
}
