package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHealthcheckHandler(t *testing.T) {
	// ***** END-TO-END TESTING *****

	app := newTestApplication(t)

	// Spin up a test server and make a GET request to it.

	// Doing this gives us a test server that exactly mimics our application
	// routes, middleware and handlers, and is a big upside of the work that we
	// did earlier to isolate all our application routing in the app.routes()
	// method.
	//
	// Notice that we defer a call to ts.Close() to shutdown the server when the
	// test finishes.
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/v1/healthcheck")

	// We can then check the value of the response status code and body.
	if code != http.StatusOK {
		t.Fatalf("want %d; got %d", http.StatusOK, code)
	}
	// JSON decoding
	var got envelope
	err := json.NewDecoder(body).Decode(&got)
	if err != nil {
		t.Fatalf("unable to parse response from server %q into envelope, '%v'", body, err)
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
