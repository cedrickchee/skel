package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedrickchee/skel/internal/data"
	"github.com/cedrickchee/skel/internal/jsonlog"
)

// Create a newTestApplication helper which returns an instance of our
// application struct containing mocked dependencies.
func newTestApplication(t *testing.T) *application {
	logger := jsonlog.New(ioutil.Discard, jsonlog.LevelInfo)
	return &application{
		logger: logger,
		models: data.NewMockModels(),
	}
}

// Define a custom testServer type which anonymously embeds a httptest.Server
// instance.
type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)
	return &testServer{ts}
}
