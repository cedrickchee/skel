package main

import (
	"net/url"
	"testing"
)

func TestReadString(t *testing.T) {
	// ***** URL PARSING EXAMPLE *****

	// u, err := url.Parse("http://example.com/search?q=golang")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// u.Scheme = "https"
	// u.Host = "google.com"
	// q := u.Query() // returns type Values map[string][]string
	// q.Set("q", "rust")
	// u.RawQuery = q.Encode()
	// t.Log(u)

	// ***** UNIT TEST *****

	// Initialize a new application object.
	var app *application

	// url, err := url.Parse("http://openmoviedb.dev/v1/movies?title=titanic")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// qs := url.Query()
	// title := app.readString(qs, "title", "")

	// if title != "titanic" {
	// 	t.Errorf("want %q; got %q", "titanic", title)
	// }

	// ***** TABLE-DRIVEN TESTS *****

	// Now expand our unit test to cover some additional test cases.
	//
	// In Go, an idiomatic way to run multiple test cases is to use table-driven
	// tests.
	tests := []struct {
		// Test case name
		name string
		// Input
		url string
		// Expected output
		want string
	}{
		{
			name: "TitleExists",
			url:  "http://openmoviedb.dev/v1/movies?title=titanic",
			want: "titanic",
		},
		{
			name: "NoTitle",
			url:  "http://openmoviedb.dev/v1/movies",
			want: "",
		},
		{
			name: "Empty",
			url:  "http://openmoviedb.dev/v1/movies?title=",
			want: "",
		},
	}

	for _, tt := range tests {
		// Use the t.Run() function to run a sub-test for each test case. The
		// first parameter to this is the name of the test (which is used to
		// identify the sub-test in any log output) and the second parameter is
		// and anonymous function containing the actual test for each case.
		t.Run(tt.name, func(t *testing.T) {
			// Initialize a new *url.URL object and pass its query string value
			// to the readString method.
			url, err := url.Parse(tt.url)
			if err != nil {
				t.Fatal(err)
			}

			// Get the url.Values map containing the query string data.
			qs := url.Query()

			// Use our helpers to extract the title query string value, falling
			// back to defaults of an empty string if they are not provided by
			// the client.
			title := app.readString(qs, "title", "")
			// Check that the output from the readString method is in the value we
			// expect. If it isn't what we expect, use the t.Errorf() function to
			// indicate that the test has failed and log the expected and actual values.
			if title != tt.want {
				t.Errorf("want %q; got %q", tt.want, title)
			}
		})
	}
}

/*
Run:

$ go test -v -run ^TestReadString$ github.com/cedrickchee/skel/cmd/api

As a side note, you can use the -failfast flag to stop the tests running after
the first failure, if you want, like so:

$ go test -failfast -v -run ^TestReadString$ github.com/cedrickchee/skel/cmd/api

To run all the tests — you can use the ./... wildcard pattern like so:
$ go test -v ./...
*/
