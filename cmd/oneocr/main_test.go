package main

import "testing"

func TestRunRejectsUnknownCommand(t *testing.T) {
	if code := run([]string{"missing"}); code != 2 {
		t.Fatalf("run unknown command code = %d, want 2", code)
	}
}
