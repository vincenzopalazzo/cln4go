package plugin

import (
	"github.com/vincenzopalazzo/cln4go/comm"
)

type getManifest[T any] struct{}

func (instance *getManifest[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	result["options"] = comm.GenerateArray(plugin.Options, map[string]bool{"getmanifest": true, "init": true})
	result["rpcmethods"] = comm.GenerateArray(plugin.RpcMethods, map[string]bool{})
	result["hooks"] = comm.GenerateArray(plugin.Hooks, map[string]bool{})
	result["subscriptions"] = comm.GenerateKeyArray(plugin.Subscriptions)
	// TODO: add notifications
	result["notifications"] = make([]string, 0)
	result["dynamic"] = plugin.dynamic
	result["featurebits"] = plugin.FeatureBits
	return result, nil
}

type initMethod[T any] struct{}

func (instance *initMethod[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	plugin.Configuration, _ = request["configuration"].(map[string]any)
	opts := request["options"].(map[string]any)
	for key, value := range opts {
		plugin.Options[key].Value = value
	}

	return plugin.onInit(plugin.State, plugin.Configuration), nil
}

func DummyOnInit[T any](state T, conf map[string]any) map[string]any {
	return map[string]any{}
}
