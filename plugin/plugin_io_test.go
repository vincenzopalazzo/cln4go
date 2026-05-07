package plugin

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
)

type ioTestState struct{}

func newIOTestPlugin() *Plugin[*ioTestState] {
	p := New[*ioTestState](&ioTestState{}, false, func(plugin *Plugin[*ioTestState], config map[string]any) map[string]any {
		return map[string]any{}
	})
	p.RegisterRPCMethod("echo", "", "", func(plugin *Plugin[*ioTestState], request Map) (Map, error) {
		return Map{"echoed": request}, nil
	})
	return p
}

// parseResponses splits the captured stdout stream on "\n\n" and unmarshals
// each chunk that is not a "log" notification into a Response. log
// notifications are returned separately so tests can also assert what the
// plugin logged.
func parseResponses(t *testing.T, raw []byte) (responses []Response, logs []map[string]any) {
	t.Helper()
	for _, chunk := range bytes.Split(raw, []byte("\n\n")) {
		if len(bytes.TrimSpace(chunk)) == 0 {
			continue
		}
		var generic map[string]any
		if err := json.Unmarshal(chunk, &generic); err != nil {
			t.Fatalf("output chunk %q is not valid JSON: %s", chunk, err)
		}
		if method, ok := generic["method"].(string); ok && method == "log" {
			if params, ok := generic["params"].(map[string]any); ok {
				logs = append(logs, params)
			}
			continue
		}
		var resp Response
		if err := json.Unmarshal(chunk, &resp); err != nil {
			t.Fatalf("response chunk %q is not valid JSON: %s", chunk, err)
		}
		responses = append(responses, resp)
	}
	return responses, logs
}

// TestRunHandlesConcatenatedRequests reproduces the failure mode that
// crashed lightningd: two JSON-RPC requests delivered in a single Read
// call. The previous buffer-size heuristic fed the concatenated bytes to
// json.Unmarshal which failed with "invalid character '{' after top-level
// value" and panicked. The scanner-based loop must split on "\n\n" and
// answer both requests independently.
func TestRunHandlesConcatenatedRequests(t *testing.T) {
	p := newIOTestPlugin()

	req1 := `{"jsonrpc":"2.0","id":1,"method":"echo","params":{"a":"foo"}}`
	req2 := `{"jsonrpc":"2.0","id":2,"method":"echo","params":{"a":"bar"}}`
	in := strings.NewReader(req1 + "\n\n" + req2 + "\n\n")
	var out bytes.Buffer

	p.run(in, &out)

	responses, _ := parseResponses(t, out.Bytes())
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d (output: %q)", len(responses), out.String())
	}
	for i, want := range []float64{1, 2} {
		gotID, ok := responses[i].Id.(float64)
		if !ok {
			t.Fatalf("response[%d] id is not a number: %T %v", i, responses[i].Id, responses[i].Id)
		}
		if gotID != want {
			t.Errorf("response[%d] id = %v, want %v", i, gotID, want)
		}
		if responses[i].Error != nil {
			t.Errorf("response[%d] returned error: %v", i, responses[i].Error)
		}
	}
}

// TestRunHandlesSplitRequest verifies that a request delivered across
// multiple Read calls (e.g. when lightningd or the OS pipe buffer flushes
// partial bytes) is reassembled into a single message.
func TestRunHandlesSplitRequest(t *testing.T) {
	p := newIOTestPlugin()

	in := &chunkedReader{
		chunks: [][]byte{
			[]byte(`{"jsonrpc":"2.0","id":42,"method":"echo",`),
			[]byte(`"params":{"x":"y"}}` + "\n\n"),
		},
	}
	var out bytes.Buffer

	p.run(in, &out)

	responses, _ := parseResponses(t, out.Bytes())
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d (output: %q)", len(responses), out.String())
	}
	gotID, ok := responses[0].Id.(float64)
	if !ok || gotID != 42 {
		t.Errorf("response id = %v, want 42", responses[0].Id)
	}
}

// TestRunRecoversFromMalformedRequest ensures a single corrupt frame
// does not take down the loop. The plugin should log the parse error,
// skip the malformed bytes, and continue processing the next request.
func TestRunRecoversFromMalformedRequest(t *testing.T) {
	p := newIOTestPlugin()

	garbage := `{not valid json{`
	good := `{"jsonrpc":"2.0","id":7,"method":"echo","params":{}}`
	in := strings.NewReader(garbage + "\n\n" + good + "\n\n")
	var out bytes.Buffer

	p.run(in, &out)

	responses, logs := parseResponses(t, out.Bytes())
	if len(responses) != 1 {
		t.Fatalf("expected 1 response after recovery, got %d (output: %q)", len(responses), out.String())
	}
	gotID, ok := responses[0].Id.(float64)
	if !ok || gotID != 7 {
		t.Errorf("response id = %v, want 7", responses[0].Id)
	}
	foundParseLog := false
	for _, log := range logs {
		if msg, _ := log["message"].(string); strings.Contains(msg, "failed to parse request") {
			foundParseLog = true
			break
		}
	}
	if !foundParseLog {
		t.Errorf("expected a 'failed to parse request' log notification, got logs: %v", logs)
	}
}

// TestWriteMessageIncludesDelimiter asserts the contract of the shared
// writer: every JSON-RPC message must be terminated with "\n\n" so the
// peer can correctly tokenise the stream.
func TestWriteMessageIncludesDelimiter(t *testing.T) {
	p := newIOTestPlugin()
	var out bytes.Buffer
	p.writer = bufio.NewWriter(&out)

	if err := p.writeMessage(map[string]any{"hello": "world"}); err != nil {
		t.Fatalf("writeMessage returned error: %s", err)
	}

	if !bytes.HasSuffix(out.Bytes(), []byte("\n\n")) {
		t.Errorf("output missing \\n\\n delimiter: %q", out.String())
	}
	stripped := bytes.TrimSuffix(out.Bytes(), []byte("\n\n"))
	if bytes.Contains(stripped, []byte("\n\n")) {
		t.Errorf("delimiter appears mid-message: %q", out.String())
	}
}

// TestLogAndResponseDoNotInterleave drives writeMessage from many
// goroutines concurrently to confirm the writerMu prevents partial
// messages from being interleaved on stdout. Without the mutex two
// concurrent writes can produce output like "{...{...}...}\n\n\n\n",
// which is what crashed the old reader loop on the peer side.
func TestLogAndResponseDoNotInterleave(t *testing.T) {
	p := newIOTestPlugin()
	var out bytes.Buffer
	p.writer = bufio.NewWriter(&out)

	const goroutines = 16
	const perGoroutine = 32
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(idx int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				if err := p.writeMessage(map[string]any{
					"goroutine": idx,
					"index":     i,
				}); err != nil {
					t.Errorf("writeMessage returned error: %s", err)
					return
				}
			}
		}(g)
	}
	wg.Wait()

	chunks := bytes.Split(bytes.TrimSuffix(out.Bytes(), []byte("\n\n")), []byte("\n\n"))
	if got := len(chunks); got != goroutines*perGoroutine {
		t.Fatalf("expected %d framed messages, got %d", goroutines*perGoroutine, got)
	}
	for i, chunk := range chunks {
		var payload map[string]any
		if err := json.Unmarshal(chunk, &payload); err != nil {
			t.Fatalf("chunk %d not valid JSON (%s): %q", i, err, chunk)
		}
	}
}

// TestRunHandlesLargeFrame confirms the new bufio.Reader-based loop has
// no fixed frame-size cap. A 1 MiB params payload — far larger than the
// 64 KB default scanner buffer that the previous bufio.Scanner-based
// implementation grew from — must be parsed and answered without
// truncating the request or terminating the loop.
func TestRunHandlesLargeFrame(t *testing.T) {
	p := newIOTestPlugin()

	big := strings.Repeat("a", 1<<20) // 1 MiB
	req := fmt.Sprintf(`{"jsonrpc":"2.0","id":7,"method":"echo","params":{"data":%q}}`, big)
	in := strings.NewReader(req + "\n\n")
	var out bytes.Buffer

	p.run(in, &out)

	responses, _ := parseResponses(t, out.Bytes())
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Errorf("response returned error: %v", responses[0].Error)
	}
	gotID, ok := responses[0].Id.(float64)
	if !ok || gotID != 7 {
		t.Errorf("response id = %v, want 7", responses[0].Id)
	}
}

// TestDispatchSendsErrorOnUnencodableResponse covers the fallback path:
// when an RPC handler returns a Map containing a value encoding/json
// cannot marshal (here, a function), dispatchRequest must still answer
// the request id with a JSON-RPC error response so the caller does not
// hang waiting for a reply that never comes.
func TestDispatchSendsErrorOnUnencodableResponse(t *testing.T) {
	p := newIOTestPlugin()
	p.RegisterRPCMethod("badreturn", "", "", func(plugin *Plugin[*ioTestState], request Map) (Map, error) {
		// Functions are not encodable by encoding/json — Marshal returns
		// "json: unsupported type: func()".
		return Map{"unencodable": func() {}}, nil
	})

	req := `{"jsonrpc":"2.0","id":99,"method":"badreturn","params":{}}`
	in := strings.NewReader(req + "\n\n")
	var out bytes.Buffer

	p.run(in, &out)

	responses, _ := parseResponses(t, out.Bytes())
	if len(responses) != 1 {
		t.Fatalf("expected 1 fallback response, got %d (output: %q)", len(responses), out.String())
	}
	if responses[0].Error == nil {
		t.Fatalf("expected error in fallback response, got: %+v", responses[0])
	}
	if responses[0].Error.Code != -32603 {
		t.Errorf("fallback error code = %d, want -32603", responses[0].Error.Code)
	}
	gotID, ok := responses[0].Id.(float64)
	if !ok || gotID != 99 {
		t.Errorf("fallback response id = %v, want 99", responses[0].Id)
	}
}

// chunkedReader hands out the configured byte chunks one Read() at a
// time so tests can simulate fragmented stdin delivery.
type chunkedReader struct {
	chunks [][]byte
	idx    int
}

func (r *chunkedReader) Read(p []byte) (int, error) {
	for r.idx < len(r.chunks) && len(r.chunks[r.idx]) == 0 {
		r.idx++
	}
	if r.idx >= len(r.chunks) {
		return 0, io.EOF
	}
	n := copy(p, r.chunks[r.idx])
	r.chunks[r.idx] = r.chunks[r.idx][n:]
	return n, nil
}
