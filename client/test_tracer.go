package client

import (
	"log"
	"testing"

	"github.com/vincenzopalazzo/cln4go/comm/tracer"
)

type TestTracer struct {
    t *testing.T
}

func (self *TestTracer) Log(level tracer.TracerLevel, msg string) {}

func (self *TestTracer) Logf(level tracer.TracerLevel, msg string, args ...any) {}

func (self *TestTracer) Info(msg string) {
	log.Println(msg)
}

func (self *TestTracer) Infof(msg string, args ...any) {
	log.Printf(msg, args...)
}
