package main

import (
	"github.com/vincenzopalazzo/cln4go/plugin"
)

type PluginStateInit struct{}

type OnRPCCommandInit[T PluginStateInit] struct{}

func (instance *OnRPCCommandInit[PluginStateInit]) Call(plugin *plugin.Plugin[PluginStateInit], request map[string]any) {
}

type HelloInit[T PluginStateInit] struct{}

func (instance *HelloInit[PluginStateInit]) Call(plugin *plugin.Plugin[PluginStateInit], request map[string]any) (map[string]any, error) {
	return map[string]any{"message": "hello from go 1.18"}, nil
}

type GetFooOptionInit[T PluginStateInit] struct{}

func (instance *GetFooOptionInit[PluginStateInit]) Call(plugin *plugin.Plugin[PluginStateInit], request map[string]any) (map[string]any, error) {
	BarValue, state := plugin.GetOpt("foo")

	if state != true {
		BarValue = ""
	}
	return map[string]any{"message": BarValue}, nil
}

func onInit(state PluginStateInit, conf map[string]any) map[string]any {
	return map[string]any{}
}

/// By module resolution this is imported inside the other plugin, but just to make the test happy
func main() {
	state := PluginStateInit{}
	plugin := plugin.New(&state, true, onInit)
	plugin.RegisterOption("foov2", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hellov2", "", "an example of rpc method", &HelloInit[PluginStateInit]{})
	plugin.RegisterRPCMethod("foo_barv2", "", "an example of rpc method", &GetFooOptionInit[PluginStateInit]{})
	plugin.RegisterNotification("rpc_command", &OnRPCCommandInit[PluginStateInit]{})
	plugin.Start()
}
