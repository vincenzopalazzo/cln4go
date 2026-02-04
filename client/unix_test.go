package client

import (
	"encoding/json"
	"os"
	"testing"
)

func TestUnixCallOne(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	if path == "" {
		t.Skip("CLN_UNIX_SOCKET not set, skipping integration test")
	}

	client, err := NewUnix(path)
	if err != nil {
		t.Fatal(err)
	}
	response, err := Call[map[string]any, map[string]any](client, "getinfo", map[string]any{})

	if err != nil {
		t.Fatal(err)
	}

	if response == nil {
		t.Fatal("getinfo returned nil response")
	}

	if response["id"] == "" {
		resp, _ := json.Marshal(response)
		t.Fatalf("response missing id field: %s", string(resp))
	}
}

type MapReq = map[string]any
type GetInfo struct {
	Network string `json:"network"`
}

func TestUnixCallOneTyped(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	if path == "" {
		t.Skip("CLN_UNIX_SOCKET not set, skipping integration test")
	}

	client, err := NewUnix(path)
	if err != nil {
		t.Fatal(err)
	}

	response, err := Call[MapReq, *GetInfo](client, "getinfo", make(map[string]any, 0))

	if err != nil {
		t.Fatal(err)
	}

	if response == nil {
		t.Fatal("getinfo returned nil response")
	}

	if response.Network == "" {
		t.Fatal("expected non-empty network field in getinfo response")
	}
}
