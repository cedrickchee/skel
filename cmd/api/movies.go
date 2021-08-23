package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cedrickchee/skel/internal/data"
)

// Add a createMovieHandler for the 'POST /v1/movies' endpoint. For now we
// simply return a plain-text placeholder response.
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an anonymous struct to hold the information that we expect to be
	// in the HTTP request body (note that the field names and types in the
	// struct are a subset of the Movie struct that we created earlier). This
	// struct will be our *target decode destination*.
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// Use the new readJSON() helper to decode the request body into the input
	// struct. If this returns an error we send the client the error message
	// along with a 400 Bad Request status code, just like before.
	//
	// Notice that when we call readJSON() we pass a *pointer* to the input
	// struct as the target decode destination.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Dump the contents of the input struct in a HTTP response.
	fmt.Fprintf(w, "%+v\n", input)
}

// Add a showMovieHandler for the 'GET /v1/movies/:id' endpoint. For now, we
// retrieve the interpolated 'id' parameter from the current URL and include it
// in a placeholder response.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// If the id is invalid, or is less than 1, we use the http.NotFound()
	// function to return a 404 Not Found response.
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
