package client

import (
	"fmt"
	"log"
	"os"
	"testing"
)

type GetInfo struct {
	id string
}

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
	request := Request{Method: "getinfo", Params: make(map[string]interface{}), Jsonrpc: "2.0", Id: 0}
	response, err := client.Call(request)
	if err != nil {
		panic(err)
	}
	if response == nil {
		panic("The get info is null, there is some problem with the client implementation")
	}
	log.Print(response)
	if request.Method != "getinfo" {
		panic("method is not a getinfo command")
	}
	// TODO: make an assertion on the part of what the request contains
}
