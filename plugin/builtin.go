package plugin

import (
	"github.com/vincenzopalazzo/cln4go/comm"
)

// GetManifest is a method to generate the manfest object needed by Core Damon
// to register the plugin. https://lightning.readthedocs.io/PLUGINS.html
type getManifest[T any] struct{}

func (instance *getManifest[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	result["options"] = comm.GenerateArray(plugin.Options, map[string]bool{})
	result["rpcmethods"] = comm.GenerateArray(plugin.RpcMethods, map[string]bool{"getmanifest": true, "init": true})
	result["hooks"] = comm.GenerateArray(plugin.Hooks, map[string]bool{})
	result["subscriptions"] = comm.GenerateKeyArray(plugin.Subscriptions)
	// FIXME: add notifications
	result["notifications"] = make([]string, 0)
	result["dynamic"] = plugin.dynamic
	result["featurebits"] = plugin.FeatureBits

	return result, nil
}

// initMethod method is called by Core Damon after the command line options has been
// parsed and the plugin has been loaded.
type initMethod[T any] struct{}

func (instance *initMethod[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	plugin.Configuration, _ = request["configuration"].(map[string]any)
	opts := request["options"]
	for key, value := range opts.(map[string]any) {
		plugin.Options[key].Value = value
	}

	return plugin.onInit(plugin, plugin.Configuration), nil
}

func DummyOnInit[T any](plugin *Plugin[T], conf map[string]any) map[string]any {
	return map[string]any{}
}
