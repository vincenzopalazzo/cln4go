package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
)

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
	onInit        *func(state T, config map[string]any) map[string]any
}

func New[T any](state *T, dynamic bool, onInit *func(state T, config map[string]any) map[string]any) *Plugin[T] {
	return &Plugin[T]{
		State:         *state,
		RpcMethods:    make(map[string]*rpcMethod[T]),
		Notifications: make(map[string]*rpcNotification[T]),
		Options:       make(map[string]*rpcOption),
		dynamic:       dynamic,
		onInit:        onInit,
	}
}

func (instance *Plugin[T]) RegisterRPCMethod(name string, usage string, description string, callback RPCCommand[T]) {
	instance.RpcMethods[name] = &rpcMethod[T]{
		Name:            name,
		Usage:           usage,
		Description:     description,
		LongDescription: description,
		callback:        callback,
	}
}

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

func (instance *Plugin[T]) RegisterNotification(name string, callback RPCEvent[T]) {
	instance.Notifications[name] = &rpcNotification[T]{
		onEvent:  name,
		callback: callback,
	}
}

func (instance *Plugin[T]) RegisterHook(name string, before []string, after []string, callback RPCCommand[T]) {
	instance.Hooks[name] = &rpcHook[T]{
		name:     name,
		before:   before,
		after:    after,
		callback: callback,
	}
}

func (instance *Plugin[T]) GetOpt(key string) (any, bool) {
	val, found := instance.Options[key]
	if !found {
		return nil, false
	}
	return val.Value, true
}

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

func (instance *Plugin[T]) handleNotification(onEvent string, request map[string]any) {
	callback, found := instance.Notifications[onEvent]
	if !found {
		panic(fmt.Sprintf("RPC notification with name %s not found", onEvent))
	}
	(*callback).Call(instance, request)
}

func (instance *Plugin[T]) configurePlugin() {
	instance.RegisterRPCMethod("getmanifest", "", "", &getManifest[T]{})
	instance.RegisterRPCMethod("init", "", "", &initMethod[T]{})
}

func (instance *Plugin[T]) Start() {
	instance.configurePlugin()
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	debug := bufio.NewWriter(os.Stderr)
	for {
		//read response
		// FIXME: move in https://github.com/LNOpenMetrics/lnmetrics.utils
		buffSize := 1024
		rawRequest := make([]byte, 0)
		for {
			recvData := make([]byte, buffSize)
			bytesResp1, err := reader.Read(recvData[:])

			if err != nil {
				panic(err)
			}
			rawRequest = append(rawRequest, recvData[:bytesResp1]...)

			if bytesResp1 < buffSize {
				break
			}
		}

		debug.Write(rawRequest)
		var request jsonrpcv2.Request
		if err := json.Unmarshal(rawRequest, &request); err != nil {
			panic(fmt.Sprintf("Error parsing request: %s input %s", err, string(rawRequest)))
		}
		if request.Id != nil {
			result, err := instance.callRPCMethod(request.Method, request.GetParams())
			var response jsonrpcv2.Response
			if err != nil {
				response = jsonrpcv2.Response{Id: request.Id, Error: map[string]any{"message": err, "code": -2}, Result: nil}
			} else {
				response = jsonrpcv2.Response{Id: request.Id, Error: nil, Result: result}
			}
			responseStr, err := json.Marshal(response)
			if err != nil {
				panic(err)
			}
			writer.Write(responseStr)
			writer.Flush()
		} else {
			instance.handleNotification(request.Method, request.GetParams())
		}
	}
}
