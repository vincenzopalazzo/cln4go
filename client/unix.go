package client

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"

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
		self.tracer.Infof("%s", err)
		panic(err)
	}
	return buf
}

func (self *UnixRPC) decodeToResponse(s []byte) *jsonrpcv2.Response[*string] {
	r := jsonrpcv2.Response[*string]{}
	if len(s) == 0 {
		return &r
	}
	self.tracer.Infof("cln4go: buffer pre dencoding %s", string(s))
	if err := self.encoder.DecodeFromBytes(s, &r); err != nil {
		self.tracer.Infof("%s", err)
	}
	return &r
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

	buffer := ""
	scanner := bufio.NewScanner(instance.socket)
	for scanner.Scan() {
		if line := scanner.Text(); len(line) > 0 {
			buffer += line
		}
	}
	resp := instance.decodeToResponse([]byte(buffer))

	if resp.Error != nil {
		return nil, fmt.Errorf("RPC error code: %s and msg: %s", resp.Error["code"], resp.Error["message"])
	}

	return resp.Result, nil
}
