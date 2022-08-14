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

func onInit(state PluginState, conf map[string]any) map[string]any {
	return map[string]any{}
}

func main() {
	state := PluginState{}
	plugin := plugin.New(&state, true, onInit)
	plugin.RegisterOption("foov2", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hellov2", "", "an example of rpc method", &Hello[PluginState]{})
	plugin.RegisterRPCMethod("foo_barv2", "", "an example of rpc method", &GetFooOption[PluginState]{})
	plugin.RegisterNotification("rpc_command", &OnRPCCommand[PluginState]{})
	plugin.Start()
}
