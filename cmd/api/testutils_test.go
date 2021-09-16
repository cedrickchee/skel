//lint:file-ignore U1000 WIP
// Test helpers
package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cedrickchee/skel/internal/data"
	"github.com/cedrickchee/skel/internal/jsonlog"
)

// Create a newTestApplication helper which returns an instance of our
// application struct containing mocked dependencies.
func newTestApplication(t *testing.T) *application {
	logger := jsonlog.New(ioutil.Discard, jsonlog.LevelInfo)
	// For now, this just contains a mock logger (which discard anything written
	// to them) and some mock models (only movie model).
	return &application{
		config: config{
			env: "test",
		},
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
	// Use the httptest.NewServer() function to create a new test server,
	// passing in the handler for the server. This starts up a HTTP server which
	// listens on a randomly-chosen port of your local machine for the duration
	// of the test.
	ts := httptest.NewServer(h)

	// Initialize a new cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	// Add the cookie jar to the client, so that response cookies are stored and
	// then sent with subsequent requests.
	ts.Client().Jar = jar

	// Disable redirect-following for the client. Essentially this function is
	// called after a 3xx response is received by the client, and returning the
	// http.ErrUseLastResponse error forces it to immediately return the
	// received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// get method makes a GET request to a given url path on the test server, and
// returns the response status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, io.ReadCloser) {
	// The network address that the test server is listening on is contained in
	// the ts.URL field. We can use this along with the ts.Client().Get() method
	// to make a GET request against the test server. This returns a
	// http.Response struct containing the response.
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, rs.Body
}

// authenticatedGet method makes a GET request to a given url path and auth
// token on the test server, and returns the response status code, headers and
// body.
func (ts *testServer) authenticatedGet(t *testing.T, token *data.Token, urlPath string) (int, http.Header, io.ReadCloser) {
	t.Helper()

	client := ts.Client()
	// req := httptest.NewRequest(http.MethodGet, ts.URL+tt.urlPath, nil)
	req, _ := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	req.Header.Set("Authorization", "Bearer "+token.Plaintext)
	rs, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, rs.Body
}

// Test assertion functions

func assertEqual(t *testing.T, a, b interface{}) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expecting values to be equal but got: '%v' and '%v'", a, b)
	}
}
