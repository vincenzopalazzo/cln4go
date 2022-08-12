package plugin

// https://github.com/golang/go/issues/46477#issuecomment-1134888278
// type RPCCallback[T any] = func(*Plugin[T], map[string]any) (map[string]any, error)
// type voidRPCCallback[T any] = func(*Plugin[T], map[string]any)

// RPCCommand
// FIXME: override the command pattern with the generic type alias when implemented
type RPCCommand[T any] interface {
	Call(*Plugin[T], map[string]any) (map[string]any, error)
}

type RPCEvent[T any] interface {
	Call(*Plugin[T], map[string]any)
}

type rpcOption struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
	Deprecated  bool   `json:"deprecated"`
	Value       any   `json:"-"`
}
