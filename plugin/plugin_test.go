package plugin

import (
	"os"
	"testing"

	cln4go "github.com/vincenzopalazzo/cln4go/client"
)

func TestCallFistMethod(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	client, err := cln4go.NewUnix(path)
	if err != nil {
		panic(err)
	}
	response, err := cln4go.Call[Map, Map](client, "hello", make(map[string]interface{}))
	if err != nil {
		panic(err)
	}

	message, found := response["message"]
	if !found {
		t.Error("The message is not found")
	}

	if message != "hello from go 1.18" {
		t.Errorf("message received %s different from expected %s", message, "hello from go 1.18")
	}
}

func TestOptionValueExist(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	client, err := cln4go.NewUnix(path)
	if err != nil {
		panic(err)
	}
	response, err := cln4go.Call[Map, Map](client, "foo_bar", make(map[string]interface{}))
	if err != nil {
		panic(err)
	}

	message := response["message"]
	name, found := response["name"]
	if !found {
		t.Error("The message or name not found in the response")
	}

	if message != "Hello Go" {
		t.Errorf("message received %s different from expected %s", message, "Hello Go")
	}

	if name != "cln4go-opt" {
		t.Errorf("name received %s different from expected %s", name, "cn4go-opt")
	}
}
