package client

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/vincenzopalazzo/cpstl/go/io/scan"

	"github.com/vincenzopalazzo/cln4go/comm/encoder"
	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
	"github.com/vincenzopalazzo/cln4go/comm/tracer"
)

// defaultTimeout is the default read/write deadline for each RPC call.
const defaultTimeout = 5 * time.Minute

type UnixRPC struct {
	socketPath string
	timeout    time.Duration
	tracer     tracer.Tracer
	encoder    encoder.JSONEncoder
}

// NewUnix creates a new UnixRPC instance. The socket path is stored
// and a fresh connection is created for each RPC call, avoiding
// stale socket state after failed or timed-out calls.
func NewUnix(path string) (*UnixRPC, error) {
	return &UnixRPC{
		socketPath: path,
		timeout:    defaultTimeout,
		tracer:     &tracer.DummyTracer{},
		encoder:    &encoder.GoEncoder{},
	}, nil
}

func (self *UnixRPC) SetTracer(tracer tracer.Tracer) {
	self.tracer = tracer
}

func (self *UnixRPC) SetEncoder(encoder encoder.JSONEncoder) {
	self.encoder = encoder
}

// SetTimeout configures the per-call read/write deadline.
// A zero or negative value disables the deadline.
func (self *UnixRPC) SetTimeout(timeout time.Duration) {
	self.timeout = timeout
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

// Call invoke a JSON RPC 2.0 method call by choosing a random id from 0 to 10000.
// A fresh Unix socket is created for each call and closed when done,
// preventing stale socket state from corrupting subsequent calls.
func Call[Req any, Resp any](client *UnixRPC, method string, data Req) (Resp, error) {
	socket, err := net.Dial("unix", client.socketPath)
	if err != nil {
		return *new(Resp), err
	}
	defer socket.Close()

	if client.timeout > 0 {
		if err := socket.SetDeadline(time.Now().Add(client.timeout)); err != nil {
			return *new(Resp), err
		}
	}

	id := fmt.Sprintf("cln4go/%d", rand.Intn(10000))
	request := jsonrpcv2.Request{
		Method:  method,
		Params:  data,
		Jsonrpc: "2.0",
		Id:      &id,
	}
	dataBytes := encodeToBytes(client, request)

	//send data
	if _, err := socket.Write(dataBytes); err != nil {
		return *new(Resp), err
	}

	// this scanner will read the buffer in one shot, so
	// there is no need to loop and append inside anther buffer
	// it is already done by the Scanner.
	var scanner scan.DynamicScanner
	if !scanner.Scan(socket) && scanner.Error() != nil {
		return *new(Resp), jsonrpcv2.MakeRPCError(-1, "scanner error", map[string]any{"error": scanner.Error()})
	}
	buffer := scanner.Bytes()

	resp, err := decodeToResponse[Resp](client, buffer)
	if err != nil {
		return *new(Resp), jsonrpcv2.MakeRPCError(-1, "decoding JSON fails, this is unexpected", map[string]any{"error": err})
	}

	if resp.Error != nil {
		return *new(Resp), resp.Error
	}

	return resp.Result, nil
}
