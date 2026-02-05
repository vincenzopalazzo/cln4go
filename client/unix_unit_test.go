package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
)

// startTestServer creates a temporary Unix socket and starts a goroutine
// that accepts connections and handles them with the provided function.
// Uses /tmp with a short name to stay within the Unix socket path length limit.
func startTestServer(t *testing.T, handler func(net.Conn)) string {
	t.Helper()
	f, err := os.CreateTemp("", "cln4go-*.sock")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path)
	t.Cleanup(func() { os.Remove(path) })

	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(fmt.Errorf("listen %s: %w", path, err))
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handler(conn)
		}
	}()
	return path
}

func TestNewUnixEmptyPath(t *testing.T) {
	_, err := NewUnix("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestCallDialFailure(t *testing.T) {
	client, err := NewUnix("/nonexistent/path/test.sock")
	if err != nil {
		t.Fatal(err)
	}
	_, err = Call[map[string]any, map[string]any](client, "getinfo", map[string]any{})
	if err == nil {
		t.Fatal("expected dial error")
	}
	if !strings.Contains(err.Error(), "connecting to unix socket") {
		t.Fatalf("expected dial context in error, got: %v", err)
	}
}

func TestCallServerCloseWithoutResponse(t *testing.T) {
	path := startTestServer(t, func(conn net.Conn) {
		// Read the request to avoid broken pipe on the client side,
		// then close without writing a response.
		buf := make([]byte, 4096)
		conn.Read(buf)
		conn.Close()
	})

	client, err := NewUnix(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Call[map[string]any, map[string]any](client, "getinfo", map[string]any{})
	if !errors.Is(err, ErrServerClosed) {
		t.Fatalf("expected ErrServerClosed, got: %v", err)
	}
}

func TestCallServerRPCError(t *testing.T) {
	path := startTestServer(t, func(conn net.Conn) {
		buf := make([]byte, 4096)
		conn.Read(buf)
		resp := `{"jsonrpc":"2.0","id":"cln4go/1","error":{"code":-32601,"message":"Unknown command"}}` + "\n\n"
		conn.Write([]byte(resp))
		conn.Close()
	})

	client, err := NewUnix(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Call[map[string]any, map[string]any](client, "nonexistent", map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
	var jrpcErr *jsonrpcv2.JSONRPCError
	if !errors.As(err, &jrpcErr) {
		t.Fatalf("expected *JSONRPCError, got %T: %v", err, err)
	}
	if jrpcErr.Code != -32601 {
		t.Fatalf("expected error code -32601, got %d", jrpcErr.Code)
	}
}

func TestCallSuccess(t *testing.T) {
	path := startTestServer(t, func(conn net.Conn) {
		buf := make([]byte, 4096)
		conn.Read(buf)
		resp := `{"jsonrpc":"2.0","id":"cln4go/1","result":{"id":"test-node-id","alias":"test"}}` + "\n\n"
		conn.Write([]byte(resp))
		conn.Close()
	})

	client, err := NewUnix(path)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Call[map[string]any, map[string]any](client, "getinfo", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "test-node-id" {
		t.Fatalf("expected id=test-node-id, got %v", result["id"])
	}
}

func TestCallSuccessTyped(t *testing.T) {
	type InfoResult struct {
		Id    string `json:"id"`
		Alias string `json:"alias"`
	}

	path := startTestServer(t, func(conn net.Conn) {
		buf := make([]byte, 4096)
		conn.Read(buf)
		resp := `{"jsonrpc":"2.0","id":"cln4go/1","result":{"id":"test-node-id","alias":"test-alias"}}` + "\n\n"
		conn.Write([]byte(resp))
		conn.Close()
	})

	client, err := NewUnix(path)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Call[map[string]any, *InfoResult](client, "getinfo", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Id != "test-node-id" {
		t.Fatalf("expected id=test-node-id, got %s", result.Id)
	}
	if result.Alias != "test-alias" {
		t.Fatalf("expected alias=test-alias, got %s", result.Alias)
	}
}
