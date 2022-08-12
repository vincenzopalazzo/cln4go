package main

import (
	"github.com/vincenzopalazzo/cln4go/plugin"
)

type PluginState struct{}

type OnRPCCommand[T PluginState] struct{}

func (instance *OnRPCCommand[PluginState]) Call(plugin *plugin.Plugin[PluginState], request map[string]any) {
}

type Hello[T PluginState] struct{}

func (instance *Hello[PluginState]) Call(plugin *plugin.Plugin[PluginState], request map[string]any) (map[string]any, error) {
	return map[string]any{"message": "hello from go 1.18"}, nil
}

type GetFooOption[T PluginState] struct{}

func (instance *GetFooOption[PluginState]) Call(plugin *plugin.Plugin[PluginState], request map[string]any) (map[string]any, error) {
	BarValue, state := plugin.GetOpt("foo")

	if state != true {
		BarValue = ""
	}
	return map[string]any{"message": BarValue}, nil
}

func main() {
	state := PluginState{}
	plugin := plugin.New(&state, true, nil)
	plugin.RegisterOption("foo", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hello", "", "an example of rpc method", &Hello[PluginState]{})
	plugin.RegisterRPCMethod("foo_bar", "", "an example of rpc method", &GetFooOption[PluginState]{})
	plugin.RegisterNotification("rpc_command", &OnRPCCommand[PluginState]{})
	plugin.Start()
}
