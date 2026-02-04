package client

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/vincenzopalazzo/cpstl/go/io/scan"

	"github.com/vincenzopalazzo/cln4go/comm/encoder"
	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
	"github.com/vincenzopalazzo/cln4go/comm/tracer"
)

// requestCounter is a monotonically increasing counter used to generate
// unique JSON-RPC request IDs across concurrent calls.
var requestCounter uint64

// ErrEmptyResponse is returned when the server sends an empty response body.
var ErrEmptyResponse = errors.New("empty response from server")

// defaultTimeout is the default read/write deadline for each RPC call.
// CLN blocking RPCs such as waitanyinvoice or waitblockheight may exceed
// this duration; use SetTimeout with a larger value or zero to disable
// the deadline for those calls.
const defaultTimeout = 5 * time.Minute

// UnixRPC is a JSON-RPC 2.0 client that communicates over a Unix domain socket.
// A fresh socket connection is created for each Call, so multiple goroutines
// may invoke Call concurrently. However, the setter methods (SetTracer,
// SetEncoder, SetTimeout) are not safe for concurrent use and must only
// be called during initialization, before any Call is made.
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
	if path == "" {
		return nil, fmt.Errorf("unix socket path must not be empty")
	}
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
// A zero or negative value disables the deadline, which is
// necessary for CLN blocking RPCs (e.g. waitanyinvoice,
// waitblockheight) that may take an unbounded amount of time.
func (self *UnixRPC) SetTimeout(timeout time.Duration) {
	self.timeout = timeout
}

func encodeToBytes[R any](client *UnixRPC, p R) ([]byte, error) {
	buf, err := client.encoder.EncodeToByte(p)
	if err != nil {
		client.tracer.Tracef("%s", err)
		return nil, err
	}
	return buf, nil
}

func decodeToResponse[R any](client *UnixRPC, s []byte) (*jsonrpcv2.Response[R], error) {
	r := jsonrpcv2.Response[R]{}
	if len(s) == 0 {
		return nil, ErrEmptyResponse
	}
	if err := client.encoder.DecodeFromBytes(s, &r); err != nil {
		client.tracer.Tracef("%s", err)
		return nil, err
	}
	return &r, nil
}

// Call invokes a JSON RPC 2.0 method call with a unique monotonic request ID.
// A fresh Unix socket is created for each call and closed when done,
// preventing stale socket state from corrupting subsequent calls.
func Call[Req any, Resp any](client *UnixRPC, method string, data Req) (Resp, error) {
	socket, err := net.Dial("unix", client.socketPath)
	if err != nil {
		return *new(Resp), err
	}
	defer func() {
		if err := socket.Close(); err != nil {
			client.tracer.Tracef("failed to close unix socket: %v", err)
		}
	}()

	if client.timeout > 0 {
		if err := socket.SetDeadline(time.Now().Add(client.timeout)); err != nil {
			return *new(Resp), err
		}
	}

	id := fmt.Sprintf("cln4go/%d", atomic.AddUint64(&requestCounter, 1))
	request := jsonrpcv2.Request{
		Method:  method,
		Params:  data,
		Jsonrpc: "2.0",
		Id:      &id,
	}
	dataBytes, err := encodeToBytes(client, request)
	if err != nil {
		return *new(Resp), fmt.Errorf("encoding JSON request: %w", err)
	}

	// send data
	if _, err := socket.Write(dataBytes); err != nil {
		return *new(Resp), err
	}

	// this scanner will read the buffer in one shot, so
	// there is no need to loop and append inside another buffer
	// it is already done by the Scanner.
	var scanner scan.DynamicScanner
	if !scanner.Scan(socket) {
		if err := scanner.Error(); err != nil {
			return *new(Resp), fmt.Errorf("reading response: %w", err)
		}
	}
	buffer := scanner.Bytes()

	resp, err := decodeToResponse[Resp](client, buffer)
	if err != nil {
		return *new(Resp), fmt.Errorf("decoding JSON response: %w", err)
	}

	if resp.Error != nil {
		return *new(Resp), resp.Error
	}

	return resp.Result, nil
}
