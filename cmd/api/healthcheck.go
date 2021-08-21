package main

import (
	"net/http"
)

// Declare a handler which writes a well-formed JSON response with information
// about the application status, operating environment and version.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in the
	// response.
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		// If there was an error, we log it and send the client a generic error
		// message.
		app.logger.Println(err)
		http.Error(w,
			"The server encountered a problem and could not process your request",
			http.StatusInternalServerError)
	}
}
