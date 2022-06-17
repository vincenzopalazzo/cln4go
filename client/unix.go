package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Request struct {
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	Jsonrpc string         `json:"jsonrpc"`
	Id      int            `json:"id"`
}

type Response struct {
	Result  map[string]any `json:result`
	Error   string         `json:"error"`
	Jsonrpc string         `json:jsonrpc`
	Id      int            `json:"id"`
}
type UnixRPC struct {
	socket net.Conn
}

func NewUnix(path string) (*UnixRPC, error) {
	socket, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	return &UnixRPC{
		socket: socket,
	}, nil
}

func EncodeToBytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes()
}

func DecodeToResponse(s []byte) *Response {
	r := Response{}
	dec := json.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&r)
	fmt.Print(r)
	if err != nil {
		log.Fatal(err)
	}
	return &r
}

func (instance UnixRPC) Call(data Request) (*Response, error) {
	//change request to bytes
	dataBytes := EncodeToBytes(data)
	log.Printf(string(dataBytes))
	//send data
	_, err := instance.socket.Write(dataBytes)

	if err != nil {
		return nil, err
	}
	//read response
	recvData := make([]byte, 1024)
	bytesResp1, err := instance.socket.Read(recvData[:])
	fmt.Print(string(recvData[:bytesResp1]))

	if err != nil {
		return nil, err
	}
	//decode response
	resp := DecodeToResponse(recvData[:bytesResp1])
	log.Print(resp)
	return resp, nil
}
