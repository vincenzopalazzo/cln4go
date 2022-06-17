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
	response, err := client.Call( "getinfo", make(map[string]interface{}))
	if err != nil {
		panic(err)
	}
	if response == nil {
		panic("The get info is null, there is some problem with the client implementation")
	}
	log.Print(response)
	// TODO: make an assertion on the part of what the request contains
}
