// tracer package implement the way to trace the cln4go implementation
// without couple the library with any other library and leave this job
// to the end developer to choose how to trace the library
package tracer

type TracerLevel int

const (
	Info TracerLevel = iota
	Debug
	Trace
	Error
)

type Tracer interface {
	/// Log a simple string without any fmt method
	Log(level TracerLevel, msg string)

	/// Logf print a generic message with the correct log level
	Logf(level TracerLevel, msg string, args ...any)

	/// Info log message at info level
	Info(msg string)

	/// Info log message at info level with fmt method
	Infof(msg string, args ...any)

	// FIXME: support other trace method
}
