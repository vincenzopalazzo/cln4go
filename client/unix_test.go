package client

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestUnixCallOne(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	if path == "" {
		err := fmt.Errorf("Unix path not exported with the CLN_UNIX_SOCKET env variable")
		panic(err)
	}

	client, err := NewUnix(path)
	if err != nil {
		panic(err)
	}
	response, err := Call[map[string]any, map[string]any](client, "getinfo", map[string]any{})

	if err != nil {
		panic(err)
	}

	if response == nil {
		panic("The get info is null, there is some problem with the client implementation")
	}

	if response["id"] == "" {
		resp, _ := json.Marshal(response)
		panic(fmt.Sprintf("Response received by the node is invalid: %s", string(resp)))
	}
}

type MapReq = map[string]any
type GetInfo struct {
	network string
}

func TestUnixCallOneTyped(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	if path == "" {
		err := fmt.Errorf("Unix path not exported with the CLN_UNIX_SOCKET env variable")
		panic(err)
	}

	client, err := NewUnix(path)
	if err != nil {
		panic(err)
	}

	response, err := Call[MapReq, *GetInfo](client, "getinfo", make(map[string]any, 0))

	if err != nil {
		panic(err)
	}

	if response == nil {
		panic("The get info is null, there is some problem with the client implementation")
	}

	if response.network == "bitcoin" {
		panic("the network should be different from the bitcoin one")
	}
}
