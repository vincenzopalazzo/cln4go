package client

import (
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
	response, err := client.Call("getinfo", make(map[string]interface{}))

	if err != nil {
		panic(err)
	}

	if response == nil {
		panic("The get info is null, there is some problem with the client implementation")
	}

	if response["id"] == "" {
		panic("Response received by the node is invalid")
	}
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
	response, err := Call[*UnixRPC, MapReq, GetInfo](client, "getinfo", make(map[string]any, 0))

	if err != nil {
		panic(err)
	}

	if response == nil {
		panic("The get info is null, there is some problem with the client implementation")
	}

	if response.Id == "" {
		panic("Response received by the node is invalid")
	}
}
