package plugin

type getManifest[T any] struct{}

func (instance *getManifest[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	result := make(map[string]any)
	// TODO: feel the getmanifest result
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
