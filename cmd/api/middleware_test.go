package main

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedrickchee/skel/internal/jsonlog"
)

// Initialize a new jsonlog.Logger.
var buf bytes.Buffer                              // log to a buffer
var logger = jsonlog.New(&buf, jsonlog.LevelInfo) // or discard the log to io.Discard

// Initialize an instance of config and application.
var cfg = config{
	env: "test",
}
var app = &application{
	config: cfg,
	logger: logger,
}

type testLoggerWriter struct {
	*httptest.ResponseRecorder
}

func (cw testLoggerWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func TestLogRequest(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := w.(http.Hijacker)
		if !ok {
			t.Errorf("http.Hijacker is unavailable on the writer. add the interface methods.")
		}
	})

	r := httptest.NewRequest("GET", "/v1/healthcheck", nil)
	w := testLoggerWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}
	handler := app.logRequest(testHandler)
	handler.ServeHTTP(w, r)

	// Check a log output.
	t.Log(buf.String())
}

func TestLogRequestReadFrom(t *testing.T) {
	data := []byte("OK")
	// Create a mock HTTP handler that we can pass to our logRequest middleware,
	// which writes a 200 status code and "OK" response body.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	})

	r := httptest.NewRequest("GET", "/", nil) // dummy http.Request
	w := httptest.NewRecorder()

	handler := app.logRequest(next)
	handler.ServeHTTP(w, r)

	t.Log(w.Body.String())
	assertEqual(t, data, w.Body.Bytes())
}

/*
Run:

$ go test -v -run ^TestLogRequest$ github.com/cedrickchee/skel/cmd/api
*/
