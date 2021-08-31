package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// *************************************************************************
	// Gracefully shutdown the running server
	// *************************************************************************
	// When we receive a `SIGINT` or `SIGTERM` signal, we instruct our server to
	// stop accepting any new HTTP requests, and give any in-flight requests a
	// "grace period" of 5 seconds to complete before the application is
	// terminated.

	// Create a shutdownError channel. We will use this to receive any errors
	// returned by the graceful Shutdown() function.
	shutdownError := make(chan error)

	// Start a background goroutine that intercept the OS signals.
	go func() {
		// Create a quit channel which carries os.Signal values.
		quit := make(chan os.Signal, 1)

		// Use signal.Notify() to listen for incoming SIGINT and SIGTERM signals
		// and relay them to the quit channel. Any other signals will not be
		// caught by signal.Notify() and will retain their default behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until a
		// signal is received.
		s := <-quit

		// Log a message to say that the signal has been caught. Notice that we
		// also call the String() method on the signal to get the signal name
		// and include it in the log entry properties.
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// Create a context with a 5-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Call Shutdown() on our server, passing in the context we just made.
		// Shutdown() will return nil if the graceful shutdown was successful,
		// or an error (which may happen because of a problem closing the
		// listeners, or because the shutdown didn't complete before the
		// 5-second context deadline is hit). We relay this return value to the
		// shutdownError channel.
		//
		// Importantly, the Shutdown() method does not wait for any background
		// tasks to complete, nor does it close hijacked long-lived connections
		// like WebSockets.
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		// Log a message to say that we're waiting for any background goroutines
		// to complete their tasks.
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})

		// Call Wait() to block until our WaitGroup counter is zero --
		// essentially blocking until the background goroutines have finished.
		// Then we return nil on the shutdownError channel, to indicate that the
		// shutdown completed without any issues.
		app.wg.Wait()
		shutdownError <- nil
	}()

	// Log a "starting server" message.
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// Start the HTTP server.
	//
	// Calling Shutdown() on our server will cause ListenAndServe() to
	// immediately return a http.ErrServerClosed error. So if we see this error,
	// it is actually a good thing and an indication that the graceful shutdown
	// has started. So we check specifically for this, only returning the error
	// if it is NOT http.ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, we wait to receive the return value from Shutdown() on the
	// shutdownError channel. If return value is an error, we know that there
	// was a problem with the graceful shutdown and we return the error.
	err = <-shutdownError
	if err != nil {
		return err
	}

	// At this point we know that the graceful shutdown completed successfully
	// and we log a "stopped server" message.
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
