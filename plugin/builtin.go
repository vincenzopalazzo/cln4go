package plugin

func generateArray[V any](mapData map[string]V) []V {
	v := make([]V, 0, len(mapData))

	for _, value := range mapData {
		v = append(v, value)
	}
	return v
}

func generateKeyArray[V any](mapData map[string]V) []string {
	k := make([]string, 0, len(mapData))

	for key, _ := range mapData {
		k = append(k, key)
	}

	return k
}

type getManifest[T any] struct {
	RpcMethods     []*rpcMethod[T]
	// Notifications  []*any
	Hooks          []*rpcHook[T]
	Subscriptions  []string
	Options        []*rpcOption
	Dynamic        bool
}

func (instance *getManifest[T]) Call(plugin *Plugin[T], request map[string]any) (any, error) {
	result := &getManifest[T]{

	}
	result.Options = generateArray(plugin.Options)
	result.RpcMethods = generateArray(plugin.RpcMethods)
	result.Hooks = generateArray(plugin.Hooks)
	result.Subscriptions= generateKeyArray(plugin.Subscriptions)
	// result.Notifications = generateArray(plugin.Notifications)
	result.Dynamic = plugin.dynamic;
	return result, nil
}

type initMethod[T any] struct{}

func (instance *initMethod[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	//TODO: parse options
	plugin.Configuration, _ = request["configuration"].(map[string]any)
	if plugin.onInit != nil {
		return (*plugin.onInit)(plugin.State, request), nil
	}
	return map[string]any{"hello": "hello from go 1.18"}, nil
}
