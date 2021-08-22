package main

import (
	"net/http"
)

// Declare a handler which writes a well-formed JSON response with information
// about the application status, operating environment and version.
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an envelope map containing the data for the response. Notice that
	// the way we've constructed this means the environment and version data
	// will now be nested under a system_info key in the JSON response.
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		// If there was an error, we log it and send the client a generic error
		// message.
		app.serverErrorResponse(w, r, err)
	}
}
