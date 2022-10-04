package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
)

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
}

func New[T any](state *T, dynamic bool, onInit func(plugin *Plugin[T], config map[string]any) map[string]any) *Plugin[T] {
	return &Plugin[T]{
		State:         *state,
		RpcMethods:    make(map[string]*rpcMethod[T]),
		Notifications: make(map[string]*rpcNotification[T]),
		Options:       make(map[string]*rpcOption),
		dynamic:       dynamic,
		onInit:        onInit,
	}
}

// Method to add a new rpc method to the plugin.
func (instance *Plugin[T]) RegisterRPCMethod(name string, usage string, description string, callback RPCCommand[T]) {
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
func (instance *Plugin[T]) RegisterNotification(name string, callback RPCEvent[T]) {
	instance.Notifications[name] = &rpcNotification[T]{
		onEvent:  name,
		callback: callback,
	}
}

// Method to add a new rpc hook to the plugin.
func (instance *Plugin[T]) RegisterHook(name string, before []string, after []string, callback RPCCommand[T]) {
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
	var notifyRequest = jsonrpcv2.Request[*string]{
		Id:      nil,
		Jsonrpc: "2.0",
		Method:  "log",
		Params:  payload,
	}
	notifyStr, err := json.Marshal(notifyRequest)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(notifyStr)
	writer.Flush()
}

// Configuring a plugin with the default rpc methods Core Lightning needs to work.
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
		var request jsonrpcv2.Request[any]
		if err := json.Unmarshal(rawRequest, &request); err != nil {
			panic(fmt.Sprintf("Error parsing request: %s input %s", err, string(rawRequest)))
		}
		if request.Id != nil {
			result, err := instance.callRPCMethod(request.Method, request.GetParams())
			var response jsonrpcv2.Response[any]
			if err != nil {
				instance.Log("broken", fmt.Sprintf("plugin generate an error: %s", err))
				response = jsonrpcv2.Response[any]{Id: request.Id, Error: map[string]any{"message": fmt.Sprintf("%s", err.Error()), "code": -2}, Result: nil}
			} else {
				response = jsonrpcv2.Response[any]{Id: request.Id, Error: nil, Result: result}
			}
			responseStr, err := json.Marshal(response)
			if err != nil {
				instance.Log("broken", fmt.Sprintf("Error marshalling response: %s", err))
				panic(err)
			}
			writer.Write(responseStr)
			writer.Flush()
		} else {
			instance.handleNotification(request.Method, request.GetParams())
		}
	}
}
