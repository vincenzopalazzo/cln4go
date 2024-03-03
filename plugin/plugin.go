package plugin

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/vincenzopalazzo/cln4go/comm/encoder"
	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
	"github.com/vincenzopalazzo/cln4go/comm/tracer"
)

type Id = jsonrpcv2.Id
type Map = map[string]any
type Request = jsonrpcv2.Request
type Response = jsonrpcv2.Response[Map]

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

func (instance *Plugin[T]) Log(level string, message string) {
	payload := map[string]any{
		"level":   level,
		"message": message,
	}
	var notifyRequest = jsonrpcv2.Request{
		Id:      nil,
		Jsonrpc: "2.0",
		Method:  "log",
		Params:  payload,
	}
	notifyStr, err := instance.encoder.EncodeToByte(notifyRequest)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(os.Stdout)
	if _, err := writer.Write(notifyStr); err != nil {
		instance.tracer.Infof("%s", err)
	}
	writer.Flush()
}

// Configuring a plugin with the default rpc methods Core Lightning needs to work.
func (self *Plugin[T]) configurePlugin() {
	self.RegisterRPCMethod("getmanifest", "", "", GetManifest[T])
	self.RegisterRPCMethod("init", "", "", InitCall[T])
}

func (self *Plugin[T]) Start() {
	self.configurePlugin()
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	for {
		//read response
		// FIXME: move in https://github.com/LNOpenMetrics/lnmetrics.utils
		buffSize := 1024
		rawRequest := make([]byte, 0)
		for {
			recvData := make([]byte, buffSize)
			bytesResp1, err := reader.Read(recvData[:])

			if err != nil {
				if err == io.EOF {
					return
				}
				panic(err)
			}
			rawRequest = append(rawRequest, recvData[:bytesResp1]...)

			if bytesResp1 < buffSize {
				break
			}
		}

		var request Request
		if err := self.encoder.DecodeFromBytes(rawRequest, &request); err != nil {
			panic(fmt.Sprintf("Error parsing request: %s input %s", err, string(rawRequest)))
		}
		if request.Id != nil {
			result, err := self.callRPCMethod(request.Method, request.GetParams())
			var response Response
			if err != nil {
				self.Log("broken", fmt.Sprintf("plugin generate an error: %s", err))
				response = Response{Id: request.Id, Jsonrpc: "2.0", Error: map[string]any{"message": err.Error(), "code": -2}, Result: nil}
			} else {
				response = Response{Id: request.Id, Jsonrpc: "2.0", Error: nil, Result: result}
			}
			responseStr, err := self.encoder.EncodeToByte(response)
			if err != nil {
				self.Log("broken", fmt.Sprintf("Error marshalling response: %s", err))
				panic(err)
			}
			if _, err := writer.Write(responseStr); err != nil {
				self.tracer.Infof("%s", err)
			}
			writer.Flush()
		} else {
			self.handleNotification(request.Method, request.GetParams())
		}
	}
}
