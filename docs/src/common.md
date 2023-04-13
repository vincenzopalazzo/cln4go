# cln4go common

The common module implement a sequence of utils that help to share code 
between other modules. In particular it implement the JSON RPC 2.0 types
and also the Tracer and encoder module.

We discussed the Custom encoder inside the [client section](./client.md), and 
in this chapter we discuss how to create a custom logger and inject it inside
the client or plugin library.

A custom logger can be defined with the following code

```golang
package trace

import (
    "github.com/LNOpenMetrics/lnmetrics.utils/log"
    "github.com/vincenzopalazzo/cln4go/comm/tracer"
)

type Tracer struct{}

func (self *Tracer) Log(lebel tracer.TracerLevel, msg string) {}

func (self *Tracer) Logf(level tracer.TracerLevel, msg string, args ...any) {}

func (self *Tracer) Info(msg string) {
    log.GetInstance().Info(msg)
}

func (self *Tracer) Infof(msg string, args ...any) {
    log.GetInstance().Infof(msg, args...)
}

func (self *Tracer) Trace(msg string) {
    log.GetInstance().Error(msg)
}

func (self *Tracer) Tracef(msg string, args ...any) {
    log.GetInstance().Errorf(msg, args...)
}
```

and then on the Client or on the Plugin, just call `SetTracer(&Tracer{})`
