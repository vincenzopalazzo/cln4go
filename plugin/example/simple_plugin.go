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

func (self *PluginState) GetName() *string {
	return &self.Name
}

func (self *PluginState) SetName(name string) {
	self.Name = name
}

type OnRPCCommand[T IPluginState] struct{}

func (instance *OnRPCCommand[IPluginState]) Call(plugin *plugin.Plugin[IPluginState], request map[string]any) {
}

type Hello[T IPluginState] struct{}

func (instance *Hello[IPluginState]) Call(plugin *plugin.Plugin[IPluginState], request map[string]any) (map[string]any, error) {
	return map[string]any{"message": "hello from go 1.18"}, nil
}

type GetFooOption[T IPluginState] struct{}

func (instance *GetFooOption[IPluginState]) Call(plugin *plugin.Plugin[IPluginState], request map[string]any) (map[string]any, error) {
	BarValue, state := plugin.GetOpt("foo")
	plugin.State.SetName("cln4go-opt")

	if state != true {
		BarValue = ""
	}
	return map[string]any{"message": BarValue, "name": plugin.GetState().GetName()}, nil
}

func main() {
	state := PluginState{
		Name: "cln4go",
	}
	plugin := plugin.New(&state, true, plugin.DummyOnInit[*PluginState])
	plugin.RegisterOption("foo", "string", "Hello Go", "An example of option", false)
	plugin.RegisterRPCMethod("hello", "", "an example of rpc method", &Hello[*PluginState]{})
	plugin.RegisterRPCMethod("foo_bar", "", "an example of rpc method", &GetFooOption[*PluginState]{})
	plugin.RegisterNotification("rpc_command", &OnRPCCommand[*PluginState]{})
	plugin.Start()
}
