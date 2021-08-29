package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of
		// a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response
				// has been sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type interface{}, so
				// we use fmt.Errorf() to normalize it into an error and call
				// our serverErrorResponse() helper. In turn, this will log the
				// error using our custom Logger type at the ERROR level and
				// send the client a 500 Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// rateLimit is an IP-based rate limiter.
// Unlike a a global rate limiter, it’s generally more common to want a separate
// rate limiter for each client, so that one bad client making too many requests
// doesn’t affect all the others.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Declare a mutex and a map to hold the clients' IP addresses and rate
	// limiters.
	var (
		mu sync.Mutex
		// An in-memory map of rate limiters, using the IP address for each
		// client as the map key.
		clients = make(map[string]*rate.Limiter)
	)

	// The function we are returning is a closure, which 'closes over' the
	// clients variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the client's IP address from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Because we’ll potentially have multiple goroutines accessing the map
		// concurrently, we’ll need to protect access to the map by using a
		// mutex to prevent race conditions.
		//
		// Lock the mutex to prevent this code from being executed concurrently.
		mu.Lock()

		// Check to see if the IP address already exists in the map. If it
		// doesn't, then initialize a new rate limiter and add the IP address
		// and limiter to the map.
		if _, found := clients[ip]; !found {
			// The rate limiter allows an average of 2 requests per second, with
			// a maximum of 4 requests in a single "burst".
			clients[ip] = rate.NewLimiter(2, 4)
		}

		// Call the Allow() method on the rate limiter for the current IP
		// address to see if the request is permitted, and if it's not, then
		// unlock the mutex and call the rateLimitExceededResponse() helper to
		// return a 429 Too Many Requests response.
		//
		// Whenever we call the Allow() method on the rate limiter exactly one
		// token will be consumed from the bucket. If there are no tokens left
		// in the bucket, then Allow() will return `false` and that acts as the
		// trigger for us send the client a `429 Too Many Requests` response.
		//
		// It’s also important to note that the code behind the Allow() method
		// is protected by a mutex and is safe for concurrent use.
		if !clients[ip].Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		// Very importantly, unlock the mutex before calling the next handler in
		// the chain. Notice that we DON'T use defer to unlock the mutex, as
		// that would mean that the mutex isn't unlocked until all the handlers
		// downstream of this middleware have also returned.
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
