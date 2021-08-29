package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func (app *application) serve() error {
	// Declare a HTTP server with some sensible timeout settings, which listens
	// on the port provided in the config struct and uses the httprouter
	// instance returned by app.routes() as the server handler.
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
		// Create a new Go log.Logger instance with the log.New() function,
		// passing in our custom Logger as the first parameter. The "" and 0
		// indicate that the log.Logger instance should not use a prefix or any
		// flags.
		ErrorLog:     log.New(app.logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Log a "starting server" message.
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Start the HTTP server, returning any error.
	return srv.ListenAndServe()
}
