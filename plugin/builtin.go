package plugin

type getManifest[T any] struct{}

func (instance *getManifest[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	return map[string]any{"hello": "hello from go 1.18"}, nil
}

func (instance *getManifest[T]) VoidCall(plugin *Plugin[T], request map[string]any) {}

type initMethod[T any] struct{}

func (instance *initMethod[T]) Call(plugin *Plugin[T], request map[string]any) (map[string]any, error) {
	return map[string]any{"hello": "hello from go 1.18"}, nil
}

func (instance *initMethod[T]) VoidCall(plugin *Plugin[T], request map[string]any) {}
