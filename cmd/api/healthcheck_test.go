package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cedrickchee/skel/internal/jsonlog"
)

func TestHealthcheckHandler(t *testing.T) {
	// ***** END-TO-END TESTING *****

	// Create a new instance of our application struct. For now, this just
	// contains a mock logger (which discard anything written to them).
	app := &application{
		config: config{
			env: "test",
		},
		logger: jsonlog.New(io.Discard, jsonlog.LevelInfo),
	}

	// We then use the httptest.NewServer() function to create a new test
	// server, passing in the value returned by our app.routes() method as the
	// handler for the server. This starts up a HTTP server which listens on a
	// randomly-chosen port of your local machine for the duration of the test.
	// Notice that we defer a call to ts.Close() to shutdown the server when the
	// test finishes.
	//
	// Doing this gives us a test server that exactly mimics our application
	// routes, middleware and handlers, and is a big upside of the work that we
	// did earlier to isolate all our application routing in the app.routes()
	// method.
	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	// The network address that the test server is listening on is contained in
	// the ts.URL field. We can use this along with the ts.Client().Get() method
	// to make a GET /healthcheck request against the test server. This returns
	// a http.Response struct containing the response.
	rs, err := ts.Client().Get(ts.URL + "/v1/healthcheck")
	if err != nil {
		t.Fatal(err)
	}

	// We can then check the value of the response status code and body.
	if rs.StatusCode != http.StatusOK {
		t.Fatalf("want %d; got %d", http.StatusOK, rs.StatusCode)
	}
	// JSON decoding
	var got envelope
	err = json.NewDecoder(rs.Body).Decode(&got)
	if err != nil {
		t.Fatalf("unable to parse response from server %q into envelope, '%v'", rs.Body, err)
	}
	env := got["system_info"].(map[string]interface{})["environment"]
	if env != "test" {
		t.Errorf("expected environment to equal %q but got %q", "test", env)
	}
}

/*
Run:

$ go test -v -run ^TestHealthcheckHandler$ github.com/cedrickchee/skel/cmd/api
*/
