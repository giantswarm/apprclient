package apprclienttest

import (
	"testing"
)

func Test_New(t *testing.T) {
	c := Config{}

	// Test that New doesn't panic and apprclient.Interface is implemented.
	_, err := New(c)
	if err != nil {
		t.Fatalf("error == %#v, want nil", err)
	}
}
