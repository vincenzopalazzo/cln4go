package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
)

type Plugin[T any] struct {
	State         *T
	rpcMethods     map[string]*rpcMethod[T]
	notifications  map[string]*rpcNotification[T]
	hooks          map[string]*rpcHook[T]
	subscriptions  map[string]*rpcNotification[T]
	Options       map[string]*rpcOption
	dynamic       bool
	Configuration map[string]any
	onInit        *func(state *T, config map[string]any) map[string]any
}

func New[T any](state *T, dynamic bool, onInit *func(state *T, config map[string]any) map[string]any) *Plugin[T] {
	return &Plugin[T]{
		State:        state,
		rpcMethod:    make(map[string]*rpcMethod[T]),
		notification: make(map[string]*rpcNotification[T]),
		Options:      make(map[string]*rpcOption),
		dynamic:      dynamic,
		onInit:       onInit,
	}
}

func (instance *Plugin[T]) RegisterRPCMethod(name string, usage string, description string, callback RPCCommand[T]) {
	instance.rpcMethod[name] = &rpcMethod[T]{
		name:            name,
		usage:           usage,
		description:     description,
		longDescription: description,
		callback:        callback,
	}
}

func (instance *Plugin[T]) RegisterNotification(name string, callback RPCEvent[T]) {
	instance.notifications[name] = &rpcNotification[T]{
		onEvent:  name,
		callback: callback,
	}
}

func (instance *Plugin[T]) RegisterHook(name string, before []string, after []string, callback RPCCommand[T]) {
	instance.hooks[name] = &rpcHook[T]{
		name: 	  	name,
		before: 	   	before,
		after: 	   	after,
		callback:  	callback,
	}
}

func (instance *Plugin[T]) RegisterSubscription(name string, callback RPCEvent[T]) {
	instance.subscriptions[name] = &rpcNotification[T]{
		onEvent:  name,
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
	callback, found := instance.rpcMethods[methodName]
	if !found {
		return nil, fmt.Errorf("RPC method with name %s not found", methodName)
	}
	return (*callback).Call(instance, request)
}

func (instance *Plugin[T]) handleNotification(onEvent string, request map[string]any) {
	callback, found := instance.notification[onEvent]
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
	reader := bufio.NewReader(os.Stdin)
	for {
		rawRequest, _, err := reader.ReadLine()
		if err != nil {
			panic(err)
		}
		var request jsonrpcv2.Request
		if err := json.Unmarshal(rawRequest, &request); err != nil {
			panic(err)
		}
		if request.Id != nil {
			result, _ := instance.callRPCMethod(request.Method, request.Params)
			response := jsonrpcv2.Response{Id: request.Id, Error: nil, Result: result}
			responseStr, err := json.Marshal(response)
			if err != nil {
				panic(err)
			}
			fmt.Print(string(responseStr))
		} else {
			instance.handleNotification(request.Method, request.Params)
		}
	}
}
