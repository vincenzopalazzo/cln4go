package client

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type GetInfo struct {
	Id string `json:"id"`
}

type MapReq = map[string]any

func TestGenericCallOne(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	if path == "" {
		err := fmt.Errorf("Unix path not exported with the CLN_UNIX_SOCKET env variable")
		panic(err)
	}

	client, err := NewUnix(path)
	if err != nil {
		panic(err)
	}
	client.SetTracer(&TestTracer{})
	response, err := Call[*UnixRPC, MapReq, GetInfo](client, "getinfo", make(map[string]interface{}))

	if err != nil {
		panic(err)
	}

	if response == nil {
		panic("The get info is null, there is some problem with the client implementation")
	}

	if response.Id == "" {
		resp, _ := json.Marshal(response)
		panic(fmt.Sprintf("response received by the node is invalid %s", string(resp)))
	}
}
