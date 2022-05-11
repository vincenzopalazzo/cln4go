package client

import (
	"fmt"
	"net"
)

type UnixRPC struct {
	socket net.Listener
}

func NewUnix(path string) (*UnixRPC, error) {
	socket, err := net.Listen("unix", "<PATH>")
	if err != nil {
		return nil, err
	}
	return &UnixRPC{
		socket: socket,
	}, nil
}

func (instance UnixRPC) Call(method string, payload map[string]any) ([]byte, error) {
	return nil, fmt.Errorf("To be implemented")
}
