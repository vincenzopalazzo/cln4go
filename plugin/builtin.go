package plugin
import (
	"encoding/json"
)

import (
	"github.com/vincenzopalazzo/cln4go/comm"
)

type getManifest[T any] struct{}

func (instance *getManifest[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	result["options"] = comm.GenerateArray(plugin.Options)
	result["rpcmethods"] = comm.GenerateArray(plugin.RpcMethods)
	result["hooks"] = comm.GenerateArray(plugin.Hooks)
	result["subscriptions"] = comm.GenerateKeyArray(plugin.Subscriptions)
	// TODO: add notifications
	result["notifications"] = make([]string, 0)
	result["dynamic"] = plugin.dynamic
	// TODO: add featurebits
	return result, nil
}

type initMethod[T any] struct{}

func (instance *initMethod[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	//TODO: parse options
	plugin.Configuration, _ = request["configuration"].(map[string]any)
	if plugin.onInit != nil {
		return (*plugin.onInit)(plugin.State, request), nil
	}
	return map[string]any{}, nil
}
