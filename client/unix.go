package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
)

type UnixRPC struct {
	socket net.Conn
}

// NewUnixRPC creates a new UnixRPC instance.
func NewUnix(path string) (*UnixRPC, error) {
	socket, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	return &UnixRPC{
		socket: socket,
	}, nil
}

func encodeToBytes(p any) []byte {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}
	return buf.Bytes()
}

// Decode the Core lightning byte response to JsonRPC
func decodeToResponse(s []byte) *jsonrpcv2.Response {
	r := jsonrpcv2.Response{}
	dec := json.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&r)
	if err != nil {
		log.Fatal(err)
	}
	return &r
}

func (instance UnixRPC) Call(method string, data map[string]any) (map[string]any, error) {
	//change request to bytes
	id := 12
	request := jsonrpcv2.Request{Method: method, Params: data, Jsonrpc: "2.0", Id: &id}
	dataBytes := encodeToBytes(request)

	//send data
	if _, err := instance.socket.Write(dataBytes); err != nil {
		return nil, err
	}

	//read response
	// FIXME: move in https://github.com/LNOpenMetrics/lnmetrics.utils
	buffSize := 1024
	buffer := make([]byte, 0)
	for {
		recvData := make([]byte, buffSize)
		bytesResp1, err := instance.socket.Read(recvData[:])

		if err != nil {
			return nil, err
		}
		buffer = append(buffer, recvData[:bytesResp1]...)

		if bytesResp1 < buffSize {
			break
		}
	}

	//decode response
	resp := decodeToResponse(buffer)

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error code: %s and msg: %s", resp.Error["code"], resp.Error["message"])
	}

	return resp.Result, nil
}
