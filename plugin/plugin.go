package plugin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/vincenzopalazzo/cln4go/comm/encoder"
	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
	"github.com/vincenzopalazzo/cln4go/comm/tracer"
)

type Id = jsonrpcv2.Id
type Map = map[string]any
type Request = jsonrpcv2.Request
type Response = jsonrpcv2.Response[Map]

// messageDelimiter terminates every JSON-RPC message exchanged with
// lightningd over the plugin protocol's stdio channel.
var messageDelimiter = []byte("\n\n")

// Plugin is the base plugin structure.
// Used to create and manage the state of a plugin.
type Plugin[T any] struct {
	State         T
	RpcMethods    map[string]*rpcMethod[T]
	Notifications map[string]*rpcNotification[T]
	Hooks         map[string]*rpcHook[T]
	Subscriptions map[string]*rpcNotification[T]
	Options       map[string]*rpcOption
	FeatureBits   map[string]any
	dynamic       bool
	Configuration map[string]any
	onInit        func(plugin *Plugin[T], config map[string]any) map[string]any
	tracer        tracer.Tracer
	encoder       encoder.JSONEncoder

	// writer is the single bufio.Writer that every JSON-RPC message
	// (responses and log notifications) is funneled through. Sharing one
	// writer across goroutines, guarded by writerMu, prevents two messages
	// from interleaving on stdout — which would otherwise produce
	// concatenated JSON that lightningd cannot parse and that would crash
	// the plugin's own read loop on the next request.
	writer   *bufio.Writer
	writerMu sync.Mutex
}

// FIXME: try to pass the pointer of the state to avoid the double copy here!
func New[T any](state T, dynamic bool, onInit func(plugin *Plugin[T], config map[string]any) map[string]any) *Plugin[T] {
	return &Plugin[T]{
		State:         state,
		RpcMethods:    make(map[string]*rpcMethod[T]),
		Notifications: make(map[string]*rpcNotification[T]),
		Options:       make(map[string]*rpcOption),
		dynamic:       dynamic,
		onInit:        onInit,
		tracer:        nil,
		encoder:       &encoder.GoEncoder{},
	}
}

func (self *Plugin[T]) SetTracer(tracer tracer.Tracer) {
	self.tracer = tracer
}

func (self *Plugin[T]) SetEncoder(encoder encoder.JSONEncoder) {
	self.encoder = encoder
}

func (self *Plugin[T]) GetEncoder() encoder.JSONEncoder {
	return self.encoder
}

func (self *Plugin[T]) GetState() T {
	return self.State
}

func (self *Plugin[T]) Encode(obj any) (map[string]any, error) {
	jsonBytes, err := self.encoder.EncodeToByte(obj)
	if err != nil {
		return nil, err
	}
	var res map[string]any
	if err := self.encoder.DecodeFromBytes(jsonBytes, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (self *Plugin[T]) Decode(payload map[string]any, destination any) error {
	bytes, err := self.encoder.EncodeToByte(payload)
	if err != nil {
		return err
	}
	return self.encoder.DecodeFromBytes(bytes, destination)
}

// Method to add a new rpc method to the plugin.
// Now supports returning a JSONRPCError for proper JSON-RPC 2.0 error propagation.
func (instance *Plugin[T]) RegisterRPCMethod(name string, usage string, description string, callback func(plugin *Plugin[T], request Map) (Map, error)) {
	instance.RpcMethods[name] = &rpcMethod[T]{
		Name:            name,
		Usage:           usage,
		Description:     description,
		LongDescription: description,
		callback:        callback,
	}
}

// Method to add a new plugin option.
func (instance *Plugin[T]) RegisterOption(name string, typ string, def string, description string, deprecated bool) {
	instance.Options[name] = &rpcOption{
		Name:        name,
		Type:        typ,
		Default:     def,
		Description: description,
		Deprecated:  deprecated,
		Value:       nil,
	}
}

// Method to add a new rpc notification to the plugin.
func (instance *Plugin[T]) RegisterNotification(name string, callback func(plugin *Plugin[T], request Map)) {
	instance.Notifications[name] = &rpcNotification[T]{
		onEvent:  name,
		callback: callback,
	}
}

// Method to add a new rpc hook to the plugin.
func (instance *Plugin[T]) RegisterHook(name string, before []string, after []string, callback func(plugin *Plugin[T], request Map) (Map, error)) {
	instance.Hooks[name] = &rpcHook[T]{
		name:     name,
		before:   before,
		after:    after,
		callback: callback,
	}
}

// Method to get a plugin option.
func (instance *Plugin[T]) GetOpt(key string) (any, bool) {
	val, found := instance.Options[key]
	if !found {
		return nil, false
	}
	return val.Value, true
}

// Method to get a plugin configuration.
func (instance *Plugin[T]) GetConf(key string) (any, bool) {
	val, found := instance.Configuration[key]
	return val, found
}

func (instance *Plugin[T]) callRPCMethod(methodName string, request map[string]any) (map[string]any, error) {
	callback, found := instance.RpcMethods[methodName]
	if !found {
		return nil, fmt.Errorf("RPC method with name %s not found", methodName)
	}
	return (*callback).Call(instance, request)
}

// Method to call notification when core lightning sends a notification.
func (instance *Plugin[T]) handleNotification(onEvent string, request map[string]any) {
	callback, found := instance.Notifications[onEvent]
	if !found {
		panic(fmt.Sprintf("RPC notification with name %s not found", onEvent))
	}
	(*callback).Call(instance, request)
}

// tracef forwards a formatted message to the configured tracer when one is
// set. Plugins built with New() default to a nil tracer, so call sites must
// not assume the tracer is available.
func (instance *Plugin[T]) tracef(format string, args ...any) {
	if instance.tracer == nil {
		return
	}
	instance.tracer.Infof(format, args...)
}

// writeMessage encodes obj as a single JSON-RPC message and writes it to the
// shared stdout writer, terminated by the "\n\n" delimiter that the CLN
// plugin protocol uses to separate messages. The write is guarded by
// writerMu so that concurrent Log() and response writes never interleave.
//
// Returns the first error encountered (encode, write, or flush). Callers
// that produce a JSON-RPC response should treat a non-nil error as a hint
// to send a fallback so the originating request id does not go unanswered.
func (instance *Plugin[T]) writeMessage(obj any) error {
	payload, err := instance.encoder.EncodeToByte(obj)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	instance.writerMu.Lock()
	defer instance.writerMu.Unlock()
	if instance.writer == nil {
		instance.writer = bufio.NewWriter(os.Stdout)
	}
	if _, err := instance.writer.Write(payload); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	if _, err := instance.writer.Write(messageDelimiter); err != nil {
		return fmt.Errorf("delimiter: %w", err)
	}
	if err := instance.writer.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}

func (instance *Plugin[T]) Log(level string, message string) {
	notifyRequest := jsonrpcv2.Request{
		Id:      nil,
		Jsonrpc: "2.0",
		Method:  "log",
		Params: map[string]any{
			"level":   level,
			"message": message,
		},
	}
	if err := instance.writeMessage(notifyRequest); err != nil {
		// Log notifications have no id, so there is no fallback we can
		// send to the peer. Surface the failure through the tracer if one
		// is configured and drop the notification — the alternative
		// (panic) would take down the plugin and, with it, lightningd.
		instance.tracef("plugin Log write error: %s", err)
	}
}

// Configuring a plugin with the default rpc methods Core Lightning needs to work.
func (self *Plugin[T]) configurePlugin() {
	self.RegisterRPCMethod("getmanifest", "", "", GetManifest[T])
	self.RegisterRPCMethod("init", "", "", InitCall[T])
}

// Start runs the plugin loop using stdin and stdout — the standard
// lightningd plugin protocol channels.
func (self *Plugin[T]) Start() {
	self.run(os.Stdin, os.Stdout)
}

// run is the testable inner loop. It reads JSON-RPC requests from in,
// dispatches them, and writes responses/log notifications to out. Splitting
// it out from Start() lets unit tests drive the protocol with arbitrary
// io.Reader/io.Writer pairs without touching the real stdio.
func (self *Plugin[T]) run(in io.Reader, out io.Writer) {
	self.configurePlugin()

	self.writerMu.Lock()
	self.writer = bufio.NewWriter(out)
	self.writerMu.Unlock()

	reader := bufio.NewReader(in)
	for {
		raw, err := readMessage(reader)
		// Process whatever bytes we received before deciding whether to
		// exit on err. This handles the case where the stream ends mid-
		// message: we still get a chance to try to parse the partial
		// payload (and log a precise error) before terminating.
		if len(bytes.TrimSpace(raw)) > 0 {
			var request Request
			if decodeErr := self.encoder.DecodeFromBytes(raw, &request); decodeErr != nil {
				// Log the parse failure and move on instead of panicking.
				// lightningd marks plugins as important and a panic here
				// would tear down the entire node — which is exactly the
				// failure mode the previous concatenated-buffer bug
				// triggered.
				self.Log("broken", fmt.Sprintf("plugin failed to parse request: %s", decodeErr))
			} else if request.Id != nil {
				self.dispatchRequest(request)
			} else {
				self.handleNotification(request.Method, request.GetParams())
			}
		}
		if err == io.EOF {
			return
		}
		if err != nil {
			self.Log("broken", fmt.Sprintf("plugin read error: %s", err))
			return
		}
	}
}

// readMessage reads bytes from r until it encounters the "\n\n" plugin
// protocol delimiter and returns the bytes up to (but not including) the
// delimiter. There is no fixed size cap — memory is bounded only by the
// size of the largest single message in the stream — so handlers that
// receive multi-megabyte RPC params or hook payloads do not get truncated
// or terminate the loop with bufio.ErrTooLong.
//
// On a clean end-of-stream between messages the function returns
// (nil, io.EOF). On EOF mid-message it returns (partial, io.EOF) so the
// caller can attempt to parse and log a precise error before exiting.
func readMessage(r *bufio.Reader) ([]byte, error) {
	var msg []byte
	for {
		line, err := r.ReadBytes('\n')
		if len(line) > 0 {
			// A bare "\n" line is the second '\n' of the "\n\n" message
			// terminator. We only treat it as a delimiter when we have
			// already accumulated some content; otherwise it is a leading
			// blank line that we silently skip.
			if len(msg) > 0 && len(line) == 1 && line[0] == '\n' {
				return bytes.TrimRight(msg, "\n"), nil
			}
			msg = append(msg, line...)
		}
		if err == nil {
			continue
		}
		if err == io.EOF {
			if len(bytes.TrimSpace(msg)) == 0 {
				return nil, io.EOF
			}
			return bytes.TrimRight(msg, "\n"), io.EOF
		}
		return msg, err
	}
}

func (self *Plugin[T]) dispatchRequest(request Request) {
	result, err := self.callRPCMethod(request.Method, request.GetParams())
	var response Response
	if err != nil {
		self.Log("broken", fmt.Sprintf("plugin generate an error: %s", err))
		// If the error is a JSONRPCError, use it directly for the JSON-RPC error object
		if jrpcErr, ok := err.(*jsonrpcv2.JSONRPCError); ok {
			response = Response{Id: request.Id, Jsonrpc: "2.0", Error: jrpcErr, Result: nil}
		} else {
			self.Log("unusual", fmt.Sprintf("plugin generate an error: %s but it is not a JSONRPCError so the plugin is losing information here!", err))
			response = Response{Id: request.Id, Jsonrpc: "2.0", Error: &jsonrpcv2.JSONRPCError{Code: -2, Message: err.Error()}, Result: nil}
		}
	} else {
		response = Response{Id: request.Id, Jsonrpc: "2.0", Error: nil, Result: result}
	}

	if writeErr := self.writeMessage(response); writeErr != nil {
		// The response we built could not be serialized (e.g. result
		// contains a channel, function, or NaN/Inf float) or the stdout
		// pipe broke. Either way, lightningd is now waiting on a reply
		// that will never arrive unless we send something. Fall back to a
		// minimal JSON-RPC error response that uses only primitive
		// values, so the request id does not go unanswered and the peer
		// does not hang.
		self.Log("broken", fmt.Sprintf("plugin failed to write response for id %v: %s", request.Id, writeErr))
		fallback := Response{
			Id:      request.Id,
			Jsonrpc: "2.0",
			Error: &jsonrpcv2.JSONRPCError{
				Code:    jsonrpcInternalError,
				Message: fmt.Sprintf("plugin response could not be serialized: %s", writeErr),
			},
			Result: nil,
		}
		if fallbackErr := self.writeMessage(fallback); fallbackErr != nil {
			// Even the fallback failed — almost certainly a broken pipe.
			// Log via the tracer (if any) and let the read loop exit on
			// the next EOF.
			self.tracef("plugin failed to write fallback response for id %v: %s", request.Id, fallbackErr)
		}
	}
}

// jsonrpcInternalError is the JSON-RPC 2.0 reserved code for "Internal
// JSON-RPC error" (https://www.jsonrpc.org/specification#error_object).
const jsonrpcInternalError = -32603
