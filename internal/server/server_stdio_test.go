package server

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"testing"
	"time"
)

func TestStartStdioReturnsOnContextCancelWhileBlockedOnStdin(t *testing.T) {
	originalStdin := os.Stdin
	originalStdout := os.Stdout

	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	t.Cleanup(func() {
		os.Stdin = originalStdin
		os.Stdout = originalStdout
		_ = stdinReader.Close()
		_ = stdinWriter.Close()
		_ = stdoutReader.Close()
		_ = stdoutWriter.Close()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := New()
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.startStdio(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case runErr := <-errCh:
		if !errors.Is(runErr, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", runErr)
		}
	case <-time.After(200 * time.Millisecond):
		_ = stdinWriter.Close()
		select {
		case <-errCh:
		case <-time.After(200 * time.Millisecond):
		}
		t.Fatal("startStdio did not return after context cancellation")
	}
}

type blockingReader struct {
	started chan struct{}
	once    sync.Once
}

func (r *blockingReader) Read(_ []byte) (int, error) {
	r.once.Do(func() {
		close(r.started)
	})
	select {}
}

func TestStartStdioWithIOReturnsOnContextCancelWhenReaderBlocks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reader := &blockingReader{started: make(chan struct{})}

	srv := New()
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.startStdioWithIO(ctx, reader, io.Discard)
	}()

	select {
	case <-reader.started:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("decoder did not start reading")
	}

	cancel()

	select {
	case runErr := <-errCh:
		if !errors.Is(runErr, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", runErr)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("startStdioWithIO did not return after context cancellation")
	}
}
