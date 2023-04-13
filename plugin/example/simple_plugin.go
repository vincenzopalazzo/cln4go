package main

import (
	"github.com/vincenzopalazzo/cln4go/plugin"
)

type IPluginState interface {
	GetName() *string
	SetName(name string)
}

type PluginState struct {
	Name string
}

func OnRpcCommand(plugin *plugin.Plugin[*PluginState], request map[string]any) {}

func Hello(plugin *plugin.Plugin[*PluginState], request map[string]any) (map[string]any, error) {
	return map[string]any{"message": "hello from go 1.18"}, nil
}

func GetOption(plugin *plugin.Plugin[*PluginState], request map[string]any) (map[string]any, error) {
	BarValue, state := plugin.GetOpt("foo")
	plugin.State.Name = "cln4go-opt"

	if !state {
		BarValue = ""
	}
	return map[string]any{"message": BarValue, "name": plugin.GetState().Name}, nil
}

func main() {
	state := PluginState{
		Name: "cln4go",
	}
	plugin := plugin.New(&state, true, plugin.DummyOnInit[*PluginState])
	plugin.RegisterOption("foo", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hello", "", "an example of rpc method", Hello)
	plugin.RegisterRPCMethod("foo_bar", "", "an example of rpc method", GetOption)
	plugin.RegisterNotification("rpc_command", OnRpcCommand)
	plugin.Start()
}
