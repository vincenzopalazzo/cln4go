// core lightning RPC benchmarks
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"

	jsonv2 "github.com/LNOpenMetrics/go-lnmetrics.reporter/pkg/json"
	"github.com/LNOpenMetrics/go-lnmetrics.reporter/pkg/model"
	"github.com/vincenzopalazzo/glightning/glightning"

	cln "github.com/vincenzopalazzo/cln4go/client"
)

// The benchmarks contains a sequence of the
// bench.
type benchResult struct {
	Bench []*bench `json:"bench"`
}

type bench struct {
	Name    string        `json:"name"`
	Runs    int           `json:"runs"`
	Times   time.Duration `json:"times"`
	TimeStr string        `json:"time_str"`
}

// runBencmarks - entry point of the benchmarks!
func runBenchmarks(benchs map[string]func(*testing.B)) {
	result := benchResult{
		Bench: []*bench{},
	}

	for name, bench_fn := range benchs {
		res := testing.Benchmark(bench_fn)
		b := bench{
			Name:    name,
			Runs:    res.N,
			Times:   res.T,
			TimeStr: res.T.String(),
		}
		log.Infof("----------------- %s -----------------", name)
		log.Infof("Number of run: %d", b.Runs)
		log.Infof("Time taken: %s", b.Times)
		result.Bench = append(result.Bench, &b)
		log.Infof("Size result: %d", len(result.Bench))
	}
	strJson, err := json.Marshal(result)
	if err != nil {
		log.Errorf("%s", err)
	}
	log.Infof("json: %s", string(strJson))
	_ = ioutil.WriteFile("bench.json", strJson, 0644)
}

var rpc *cln.UnixRPC
var gln *glightning.Lightning

func init() {
	path := os.Getenv("CLN_UNIX_SOCKET")
	if path == "" {
		err := fmt.Errorf("Unix path not exported with the CLN_UNIX_SOCKET env variable")
		panic(err)
	}

	rpc, _ = cln.NewUnix(path)
	rpc.SetEncoder(&jsonv2.FastJSON{})

	gln = glightning.NewLightning()
	url := strings.Split(path, "/")
	socket := url[len(url)-1]
	subPath := url[0 : len(url)-1]
	pathStr := strings.Join(subPath, "/")
	gln.StartUp(socket, pathStr)
}

func main() {
	bench := map[string]func(*testing.B){
		"cln4go-listnodes":        cln4goListNodes,
		"glightning-listnodes":    glightningListNodes,
		"cln4go-listchannels":     cln4goListChannels,
		"glightning-listchannels": glightningListChannels,
	}
	runBenchmarks(bench)
}

// plain call to list nodes
func cln4goListNodes(b *testing.B) {
	_, err := cln.Call[map[string]any, model.ListNodesResp](rpc, "listnodes", map[string]any{})
	if err != nil {
		log.Error("cln4go listnodes", err)
	}
}

func glightningListNodes(b *testing.B) {
	_, err := gln.ListNodes()
	if err != nil {
		log.Error("glightning listnodes", err)
	}
}

// plain call to list nodes
func cln4goListChannels(b *testing.B) {
	_, err := cln.Call[map[string]any, *model.ListChannelsResp](rpc, "listchannels", map[string]any{})
	if err != nil {
		log.Error("cln4go listchannels", err)
	}
}

func glightningListChannels(b *testing.B) {
	_, err := gln.ListChannels()
	if err != nil {
		log.Error("glightning listchannels", err)
	}
}
