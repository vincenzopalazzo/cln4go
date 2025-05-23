package client

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/vincenzopalazzo/cpstl/go/io/scan"

	"github.com/vincenzopalazzo/cln4go/comm/encoder"
	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
	"github.com/vincenzopalazzo/cln4go/comm/tracer"
)

type UnixRPC struct {
	socket  net.Conn
	tracer  tracer.Tracer
	encoder encoder.JSONEncoder
}

// NewUnixRPC creates a new UnixRPC instance.
func NewUnix(path string) (*UnixRPC, error) {
	socket, err := net.Dial("unix", path)
	if err != nil {
		return nil, err
	}
	return &UnixRPC{
		socket:  socket,
		tracer:  &tracer.DummyTracer{},
		encoder: &encoder.GoEncoder{},
	}, nil
}

func (self *UnixRPC) SetTracer(tracer tracer.Tracer) {
	self.tracer = tracer
}

func (self *UnixRPC) SetEncoder(encoder encoder.JSONEncoder) {
	self.encoder = encoder
}

func encodeToBytes[R any](client *UnixRPC, p R) []byte {
	buf, err := client.encoder.EncodeToByte(p)
	if err != nil {
		client.tracer.Tracef("%s", err)
		panic(err)
	}
	return buf
}

func decodeToResponse[R any](client *UnixRPC, s []byte) (*jsonrpcv2.Response[R], error) {
	r := jsonrpcv2.Response[R]{}
	if len(s) == 0 {
		return &r, nil
	}
	if err := client.encoder.DecodeFromBytes(s, &r); err != nil {
		client.tracer.Tracef("%s", err)
		return nil, err
	}
	return &r, nil
}

// Call invoke a JSON RPC 2.0 method call by choosing a random id from 0 to 10000
func Call[Req any, Resp any](client *UnixRPC, method string, data Req) (Resp, error) {
	id := fmt.Sprintf("cln4go/%d", rand.Intn(10000))
	request := jsonrpcv2.Request{
		Method:  method,
		Params:  data,
		Jsonrpc: "2.0",
		Id:      &id,
	}
	dataBytes := encodeToBytes(client, request)

	//send data
	if _, err := client.socket.Write(dataBytes); err != nil {
		return *new(Resp), err
	}

	// this scanner will read the buffer in one shot, so
	// there is no need to loop and append inside anther buffer
	// it is already done by the Scanner.
	var scanner scan.DynamicScanner
	if !scanner.Scan(client.socket) && scanner.Error() != nil {
		return *new(Resp), fmt.Errorf("scanner error: %s", scanner.Error())
	}
	buffer := scanner.Bytes()

	resp, err := decodeToResponse[Resp](client, buffer)
	if err != nil {
		return *new(Resp), fmt.Errorf("decoding JSON fails, this is unexpected %s", err)
	}

	if resp.Error != nil {
		return *new(Resp), fmt.Errorf("RPC error code: %d and msg: %s", resp.Error.Code, resp.Error.Message)
	}

	return resp.Result, nil
}
