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

func (self *UnixRPC) encodeToBytes(p any) []byte {
	buf, err := self.encoder.EncodeToByte(p)
	if err != nil {
		self.tracer.Tracef("%s", err)
		panic(err)
	}
	return buf
}

func (self *UnixRPC) decodeToResponse(s []byte) (*jsonrpcv2.Response[*string], error) {
	r := jsonrpcv2.Response[*string]{}
	if len(s) == 0 {
		return &r, nil
	}
	if err := self.encoder.DecodeFromBytes(s, &r); err != nil {
		self.tracer.Tracef("%s", err)
		return nil, err
	}
	return &r, nil
}

// Call invoke a JSON RPC 2.0 method call by choosing a random id from 0 to 10000
func (instance UnixRPC) Call(method string, data map[string]any) (map[string]any, error) {
	id := fmt.Sprintf("%d", rand.Intn(10000))
	request := jsonrpcv2.Request[*string]{
		Method:  method,
		Params:  data,
		Jsonrpc: "2.0",
		Id:      &id,
	}
	dataBytes := instance.encodeToBytes(request)

	//send data
	if _, err := instance.socket.Write(dataBytes); err != nil {
		return nil, err
	}

	// this scanner will read the buffer in one shot, so
	// there is no need to loop and append inside anther buffer
	// it is already done by the Scanner.
	var scanner scan.DynamicScanner
	if !scanner.Scan(instance.socket) && scanner.Error() != nil {
		return nil, fmt.Errorf("scanner error: %s", scanner.Error())
	}
	buffer := scanner.Bytes()

	resp, err := instance.decodeToResponse(buffer)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON fails, this is unexpected %s", err)
	}

	if resp.Error != nil {
		code := int64(resp.Error["code"].(float64))
		return nil, fmt.Errorf("RPC error code: %d and msg: %s", code, resp.Error["message"])
	}

	return resp.Result, nil
}
