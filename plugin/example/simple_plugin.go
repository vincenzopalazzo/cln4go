package main

import (
	"github.com/vincenzopalazzo/cln4go/comm/jsonrpcv2"
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

// Example of an RPC method returning a JSONRPCError for JSON-RPC 2.0 error propagation.
// Note: If you get a 'plugin.JSONRPCError is not a type' error, ensure your import path is correct and
// that JSONRPCError is exported from the plugin package (it is, as of the latest code).
func ErrorExample(p *plugin.Plugin[*PluginState], request map[string]any) (map[string]any, error) {
	return nil, jsonrpcv2.MakeRPCError(1001, "This is a JSON-RPC error from Go", map[string]any{"hint": "You can add extra error data here"})
}

func main() {
	state := PluginState{
		Name: "cln4go",
	}
	plugin := plugin.New(&state, true, plugin.DummyOnInit[*PluginState])
	plugin.RegisterOption("foo", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hello", "", "an example of rpc method", Hello)
	plugin.RegisterRPCMethod("foo_bar", "", "an example of rpc method", GetOption)
	plugin.RegisterRPCMethod("error_example", "", "an example of JSON-RPC error return", ErrorExample)
	plugin.RegisterNotification("rpc_command", OnRpcCommand)
	plugin.Start()
}
