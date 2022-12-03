package tracer

type DummyTracer struct{}

func (self *DummyTracer) Log(level TracerLevel, msg string) {}

func (self *DummyTracer) Logf(level TracerLevel, msg string, args ...any) {}

func (self *DummyTracer) Info(msg string) {}

func (self *DummyTracer) Infof(msg string, args ...any) {}

func (self *DummyTracer) Trace(msg string) {}

func (self *DummyTracer) Tracef(msg string, args ...any) {}
