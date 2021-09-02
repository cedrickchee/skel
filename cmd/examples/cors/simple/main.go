package main

import (
	"flag"
	"log"
	"net/http"
)

// A string constant containing the HTML for the webpage. This consists of a
// <h1> header tag, and some JavaScript which fetches the JSON from our GET
// /v1/healthcheck endpoint and writes it to inside the <div id='output'></div>
// element.
const html = `
<!DOCTYPE html>
<html lang='en'>
<head>
    <meta charset='UTF-8'>
</head>
<body>
    <h1>Simple CORS</h1>
    <div id='output'></div>
    <script>
        document.addEventListener('DOMContentLoaded', function() {
            fetch('http://localhost:4000/v1/healthcheck').then(
                function (response) {
                    response.text().then(function (text) {
                        document.getElementById('output').innerHTML = text;
                    });
                },
                function(err) {
                    document.getElementById('output').innerHTML = err;
                }
            );
        });
    </script>
</body>
</html>`

func main() {
	// Make the server address configurable at runtime via a command-line flag.
	addr := flag.String("addr", ":9000", "Server address")
	flag.Parse()

	log.Printf("starting server on %s", *addr)

	// Start a HTTP server listening on the given address, which responds to all
	// requests with the webpage HTML above.
	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
