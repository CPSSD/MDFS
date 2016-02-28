package utils_test

import (
	"github.com/CPSSD/MDFS/utils"
	"testing"
)

func TestHash(t *testing.T) {
	var tests = []struct {
		path, filename string
		want           bool
	}{
		{utils.GetUserHome() + "/.testing_files/", "test.txt", true},
		{utils.GetUserHome() + "/.testing_files/", "test.jpg", true},
		{utils.GetUserHome() + "/.testing_files/", "testing", false},
	}
	for _, c := range tests {
		got := utils.CheckForHash(c.path, c.filename)
		if got != c.want {
			t.Errorf("CheckForHash(%q, %q) == %t, want %t", c.path, c.filename, got, c.want)
		}
	}
}
