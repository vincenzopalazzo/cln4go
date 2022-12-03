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
	self.tracer.Tracef("cln4go: buffer pre dencoding %s", string(s))
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

	buffer := []byte{}
	scanner := bufio.NewScanner(instance.socket)
	// CLN return a really big buffer whe there is not filtering
	// option active, so we need a way to say. Please read till
	// the end.
	//
	// The actual what that this is implement is the more clean
	// and easy way, but there case like https://github.com/LNOpenMetrics/go-lnmetrics.reporter/issues/123
	// where we reach the max buffer size and the
	// scan abort with an invalid json.
	scanner.Buffer(buffer, bufio.MaxScanTokenSize*4)
	for scanner.Scan() {
		if line := scanner.Bytes(); len(line) > 0 {
			instance.tracer.Trace(string(line))
			buffer = append(buffer, line...)
		} else {
			break
		}
	}

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
