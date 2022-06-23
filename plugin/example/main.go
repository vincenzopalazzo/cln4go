package main

import (
	"github.com/vincenzopalazzo/cln4go/plugin"
)

type PluginState struct{}

func Hello(plugin *plugin.Plugin[PluginState], request map[string]any) (map[string]any, error) {
	return map[string]any{"hello": "hello from go 1.18"}, nil
}

func OnRPCCommand(plugin *plugin.Plugin[PluginState], request map[string]any) {}

func main() {
	state := PluginState{}
	plugin := plugin.New[PluginState](&state, false)
	plugin.AddOption("foo", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hello", "", "an example of rpc method", Hello)
	plugin.RegisterNotification("rpc_command", OnRPCCommand)
	plugin.Start()
}
