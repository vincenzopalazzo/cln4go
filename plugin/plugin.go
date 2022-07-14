package plugin

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Plugin[T any] struct {
	State         *T
	rpcMethod     map[string]*rpcMethod[T]
	notification  map[string]*rpcNotification[T]
	Options       map[string]*rpcOption
	dynamic       bool
	Configuration map[string]any
}

// TODO: add onInit callback
func New[T any](state *T, dynamic bool) *Plugin[T] {
	return &Plugin[T]{
		State:        state,
		rpcMethod:    make(map[string]*rpcMethod[T]),
		notification: make(map[string]*rpcNotification[T]),
		Options:      make(map[string]*rpcOption),
		dynamic:      dynamic,
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

func (instance *Plugin[T]) RegisterNotification(name string, callback RPCCommand[T]) {
	instance.notification[name] = &rpcNotification[T]{
		onEvent:  name,
		callback: callback,
	}
}

func (instance *Plugin[T]) AddOption(name string, typ string, defaultValue string, description string, deprecated bool) {
	instance.Options[name] = &rpcOption{
		Name:        name,
		Type:        typ,
		Default:     defaultValue,
		Description: description,
		Deprecated:  deprecated,
	}
}

func (instance *Plugin[T]) callRPCMethod(methodName string, request map[string]any) (map[string]any, error) {
	callback, found := instance.rpcMethod[methodName]
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
		var request request
		if err := json.Unmarshal(rawRequest, &request); err != nil {
			panic(err)
		}
		if request.Id != nil {
			result, _ := instance.callRPCMethod(request.Method, request.Params)
			response := response{Id: request.Id, Error: nil, Result: result}
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
