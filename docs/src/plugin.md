# cln4go Plugin

Implementing a plugin in Go it's really straightforward, and the cln4go provide
an easy interface to implement one from scratch

An hello world plugin can be implemented with

```golang
package main

import (
    "github.com/vincenzopalazzo/cln4go/plugin"
)

// Common interface of the State!
type IPluginState interface {
    GetName() *string
    SetName(name string)
}

// PluginState - each plugin has a state that is stored
// inside the plugin and mutated across the plugin lifecycle.
//
// This is the perfect place where store the UNIX client if
// needed by the plugin, otherwise the plugin do not include this
// dependencies.
type PluginState struct {
    Name string
}

func (self *PluginState) GetName() *string {
    return &self.Name
}

func (self *PluginState) SetName(name string) {
    self.Name = name
}

// A callback in cln4go is implemeted with the Command,
// So this is just a struct to implement the interface on top
type OnRPCCommand[T IPluginState] struct{}

// Implementing the callback, now the void return here, this mean that it is a notification
func (instance *OnRPCCommand[IPluginState]) Call(plugin *plugin.Plugin[IPluginState], request map[string]any) {}

// Implementing another callback
type Hello[T IPluginState] struct{}

// Implementing the callback, please note that this is not an void method, so this callback can be register
// as RPC method or as an hook.
func (instance *Hello[IPluginState]) Call(plugin *plugin.Plugin[IPluginState], request map[string]any) (map[string]any, error) {
    return map[string]any{"message": "hello from go 1.18"}, nil
}

func main() {
    state := PluginState{
        Name: "cln4go",
    }
    plugin := plugin.New(&state, true, plugin.DummyOnInit[*PluginState])
    plugin.RegisterOption("foo", "string", "Hello Go", "An example of option", false)
    plugin.RegisterRPCMethod("hello", "", "an example of rpc method", &Hello[*PluginState]{})
    plugin.RegisterNotification("rpc_command", &OnRPCCommand[*PluginState]{})
    plugin.Start()
}
```

The code to write a callback is to much, and this would be better to have just a func declaration,
so the API to define a callback can be improved but for now we leave this as it is because with generics
we have some limitation, and a feature from the Go lang side is required. We keep track of this feature
with the issue [#12](https://github.com/vincenzopalazzo/cln4go/issues/12)

## Intercept on init callback

While the API is minimal and do not include the RPC API for core lightning, with the plugin
it is possible put the RPC client inside the State and create the client inside the on init callback.

It is possible register the on init callback with a simple callback like the following code

```golang
package main

import (
    "github.com/vincenzopalazzo/cln4go/plugin"
    cln "github.com/vincenzopalazzo/cln4go/client"
)

type State struct {
    Client *cln.UnixRPC
}

func OnInit[T State](plugin *plugin.Plugin[T], request map[string]any) map[string]any {
    state := plugin.State()

    lightningDir, _ := plugin.GetConf("lightning-dir")
    rpcFile, _ := plugin.GetConf("rpc-file")

    rpcPath := strings.Join([]string{lightningDir.(string), rpcFile.(string)}, "/")
    rpc, err = cln.NewUnix(path)
    if err != nil {
        panic(err)
    }
    state.Client = rpc
    return map[string]any{}
}

func main() {
    state := State{}
    plugin := plugin.New(&state, true, OnInit[*PluginState])
    plugin.Start()
}
```

As in the [client](./client.md) it is possible define a custom encoder and set it as we did inside
the client section, but also it is register and access to a custom tracer as discussed in the [common](./common.md)
section to have a different way to log the plugin.

## Returning JSON-RPC 2.0 Errors from Go

cln4go now supports returning proper JSON-RPC 2.0 errors from your plugin methods in an idiomatic Go way. You can return a `*plugin.JSONRPCError` from any RPC method, and it will be serialized as a JSON-RPC error object on the wire, preserving the error code, message, and optional data fields.

### Example

```go
import "github.com/vincenzopalazzo/cln4go/plugin"

func ErrorExample(plugin *plugin.Plugin[*PluginState], request map[string]any) (map[string]any, error) {
    return nil, &plugin.JSONRPCError{
        Code:    1001,
        Message: "This is a JSON-RPC error from Go",
        Data:    map[string]any{"hint": "You can add extra error data here"},
    }
}
```

Register this method as usual:

```go
plugin.RegisterRPCMethod("error_example", "", "an example of JSON-RPC error return", ErrorExample)
```

When this method returns an error, the plugin framework will send a JSON-RPC error object to the client, fully compatible with the JSON-RPC 2.0 specification. If you return a regular Go error, a generic error object will be sent instead.

This makes it easy to propagate rich, structured errors from your Go plugin to any JSON-RPC client.

## Plugin Template

If you do not want start from scratch, you can use the [plugin template](https://github.com/coffee-tools/cln4go.plugin).
