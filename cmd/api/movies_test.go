package main

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/cedrickchee/skel/internal/data"
)

func TestShowMovieHandler(t *testing.T) {
	// Create a new instance of our application struct which uses the mocked
	// dependencies.
	app := newTestApplication(t)

	// Establish a new test server for running end-to-end tests.
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	notFoundErrMsg := "the requested resource could not be found"

	// Set up some table-driven tests to check the responses sent by our
	// application for different URLs.
	tests := []struct {
		name       string
		urlPath    string
		wantCode   int
		wantTitle  string
		wantErrMsg string
	}{
		{"Valid ID", "/v1/movies/1", http.StatusOK, "Casablanca", ""},
		{"Non-existent ID", "/v1/movies/2", http.StatusNotFound, "", notFoundErrMsg},
		{"Negative ID", "/v1/movies/-1", http.StatusNotFound, "", notFoundErrMsg},
		{"Decimal ID", "/v1/movies/1.23", http.StatusNotFound, "", notFoundErrMsg},
		{"String ID", "/v1/movies/avengers", http.StatusNotFound, "", notFoundErrMsg},
		{"Empty ID", "/v1/movies/", http.StatusMovedPermanently, "", ""},
		{"Trailing slash", "/v1/movies/1/", http.StatusMovedPermanently, "", ""},
	}

	user, err := app.models.Users.GetByEmail("john@example.com")
	if err != nil {
		t.Fatal(err)
	}
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, header, body := ts.authenticatedGet(t, token, tt.urlPath)

			// We can then check the value of the response status code and body.
			if code != tt.wantCode {
				t.Fatalf("want %d; got %d", tt.wantCode, code)
			}

			if header.Get("Content-Type") == "application/json" {
				// JSON decoding
				var got envelope
				err = json.NewDecoder(body).Decode(&got)
				if err != nil {
					t.Fatalf("unable to parse response from server %q into envelope, '%v'", body, err)
				}

				// Assertions
				switch code {
				case http.StatusOK:
					title := got["movie"].(map[string]interface{})["title"]
					if title != tt.wantTitle {
						t.Errorf("expected movie title to equal %q but got %q", tt.wantTitle, title)
					}
				case http.StatusNotFound:
					errMsg := got["error"].(string)
					if errMsg != tt.wantErrMsg {
						t.Errorf("expected error message to equal %q but got %q", tt.wantErrMsg, errMsg)
					}
				}
			}
		})
	}
}

/*
Run:

$ go test -v -run ^TestShowMovieHandler$ github.com/cedrickchee/skel/cmd/api
*/
