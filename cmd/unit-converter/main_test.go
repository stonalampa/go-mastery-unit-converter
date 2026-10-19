package main

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

func TestRunRejectsUnknownFlag(t *testing.T) {
	var stderr bytes.Buffer

	if err := run(context.Background(), []string{"-port", "80"}, &stderr); err == nil {
		t.Error("run returned nil, want an error for an unknown flag")
	}
}

func TestRunShutsDownCleanlyWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var stderr bytes.Buffer
	done := make(chan error, 1)

	go func() { done <- run(ctx, []string{"-addr", "127.0.0.1:0"}, &stderr) }()
	cancel() // what Ctrl-C does in main

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("run returned %v, want nil after a clean shutdown", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not return within 5s of cancellation")
	}
}

func TestRunReportsAddressInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	var stderr bytes.Buffer

	err = run(context.Background(), []string{"-addr", ln.Addr().String()}, &stderr)

	if err == nil {
		t.Error("run returned nil, want an error for an address already in use")
	}
}
