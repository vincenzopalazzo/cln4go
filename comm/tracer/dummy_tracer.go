package tracer

type DummyTracer struct{}

func (self *DummyTracer) Log(msg string) {}

func (self *DummyTracer) Logf(msg string, args ...any) {}

func (self *DummyTracer) Info(msg string) {}

func (self *DummyTracer) Infof(msg string, args ...any) {}
