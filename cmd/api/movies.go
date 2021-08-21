package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Add a createMovieHandler for the 'POST /v1/movies' endpoint. For now we
// simply return a plain-text placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// Add a showMovieHandler for the 'GET /v1/movies/:id' endpoint. For now, we
// retrieve the interpolated 'id' parameter from the current URL and include it
// in a placeholder response.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// If the id is invalid, or is less than 1, we use the http.NotFound()
	// function to return a 404 Not Found response.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}

func (app *application) exampleHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"hello": "world",
	}

	// Set the 'Content-Type: application/json' header on the response.
	w.Header().Set("Content-Type", "application/json")

	// Use the json.NewEncoder() function to initialize a json.Encoder instance
	// that writes to the http.ResponseWriter. Then we call its Encode() method,
	// passing in the data that we want to encode to JSON (which in this case is
	// the map above). If the data can be successfully encoded to JSON, it will
	// then be written to our http.ResponseWriter.
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		app.logger.Println(err)
		http.Error(w,
			"The server encountered a problem and could not process your request",
			http.StatusInternalServerError)
	}
}
