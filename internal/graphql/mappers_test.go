package graphql

import "testing"

type boomErr struct{}

func (boomErr) Error() string { return "boom" }

func TestErrString(t *testing.T) {
	if got := errString(nil); got != nil {
		t.Fatalf("nil: got %v want nil", got)
	}

	got := errString(boomErr{})
	if got == nil {
		t.Fatal("non-nil err: got nil")
	}
	if *got != "boom" {
		t.Fatalf("non-nil err: got %q want boom", *got)
	}
}
