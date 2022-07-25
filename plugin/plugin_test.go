package plugin

import (
	"os"
	"testing"

	"github.com/vincenzopalazzo/cln4go/client"
)

func TestCallFistMethod(t *testing.T) {
	path := os.Getenv("CLN_UNIX_SOCKET")
	client, err := client.NewUnix(path)
	if err != nil {
		panic(err)
	}
	response, err := client.Call("hello", make(map[string]interface{}))
	message, found := response["message"]
	if !found {
		t.Error("The message is not found")
	}

	if message != "hello from go 1.18" {
		t.Error("message received %s different from expected %s", message, "hello from go 1.18")
	}
}
