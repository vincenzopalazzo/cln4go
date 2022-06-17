package client

import (
	"log"
	"os"
	"testing"
)

type GetInfo struct {
	id string
}

func TestUnixCallOne(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
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
	log.Print(request.Method)
	if request.Method != "getinfo" {
		panic("method is not a getinfo command")
	}
	// TODO: make an assertion on the part of what the request contains
}
