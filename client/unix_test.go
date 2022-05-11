package client

import (
	"testing"
)

type GetInfo struct {
	id string
}

func TestUnixCallOne(t *testing.T) {
	// TODO take the path from the os environment
	client, err := NewUnix("<PATH>")
	if err != nil {
		panic(err)
	}
	getInfo, err := Call[*UnixRPC, GetInfo](client, "getinfo", make(map[string]any))
	if err != nil {
		panic(err)
	}
	if getInfo.id == "" {
		panic("")
	}
}
