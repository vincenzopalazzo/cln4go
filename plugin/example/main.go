package main

import (
	"github.com/vincenzopalazzo/cln4go/plugin"
)

type PluginState struct{}

type OnRPCCommand[T PluginState] struct{}

func (instance *OnRPCCommand[T]) Call(plugin *plugin.Plugin[T], request map[string]any) {}

type Hello[T PluginState] struct{}

func (instance *Hello[T]) Call(plugin *plugin.Plugin[T], request map[string]any) (map[string]any, error) {
	return map[string]any{"hello": "hello from go 1.18"}, nil
}

func main() {
	state := PluginState{}
	plugin := plugin.New(&state, false, nil)
	plugin.AddOption("foo", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hello", "", "an example of rpc method", &Hello[PluginState]{})
	plugin.RegisterNotification("rpc_command", &OnRPCCommand[PluginState]{})
	plugin.Start()
}
