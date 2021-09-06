package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cedrickchee/skel/internal/data"
	"github.com/cedrickchee/skel/internal/validator"
	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
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
//
// Using this pattern for rate-limiting will only work if your API application
// is running on a single-machine. If your infrastructure is distributed, with
// your application running on multiple servers behind a load balancer, then
// you'll need to use an alternative approach.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Define a client struct to hold the rate limiter and last seen time for
	// each client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the clients' IP addresses and rate
	// limiters.
	var (
		mu sync.Mutex
		// An in-memory map of rate limiters, using the IP address for each
		// client as the map key.
		// The map values are pointers to a client struct.
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients
	// map once every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter checks from happening
			// while the cleanup is taking place.
			mu.Lock()

			// Loop through all clients. If they haven't been seen within the
			// last three minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Importantly, unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	// The function we are returning is a closure, which 'closes over' the
	// clients variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only carry out the check if rate limiting is enabled.
		if app.config.limiter.enabled {
			// Get the client's real IP address from the request.
			ip := realip.FromRequest(r)

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
				// Create and add a new client struct to the map if it doesn't
				// already exist.
				clients[ip] = &client{
					// The rate limiter allows an average of 2 requests per second,
					// with a maximum of 4 requests in a single "burst".
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			// Update the last seen time for the client.
			clients[ip].lastSeen = time.Now()

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
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Very importantly, unlock the mutex before calling the next handler in
			// the chain. Notice that we DON'T use defer to unlock the mutex, as
			// that would mean that the mutex isn't unlocked until all the handlers
			// downstream of this middleware have also returned.
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

// authenticate checks the authentication token to authenticate users, so the
// app knows which user a request is coming from.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the 'Vary: Authorization' header to the response. This indicates
		// to any caches that the response may vary based on the value of the
		// Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request. This
		// will return the empty string "" if there is no such header found.
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization header found, use the contextSetUser()
		// helper that we just made to add the AnonymousUser to the request
		// context. Then we call the next handler in the chain and return
		// without executing any of the code below.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Otherwise, we expect the value of the Authorization header to be in
		// the format 'Bearer <token>'. We try to split this into its
		// constituent parts, and if the header isn't in the expected format we
		// return a 401 Unauthorized response using the
		// invalidAuthenticationTokenResponse() helper.
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Extract the actual authentication token from the header parts.
		token := headerParts[1]

		// Validate the token to make sure it is in a sensible format.
		v := validator.New()

		// If the token isn't valid, use the
		// invalidAuthenticationTokenResponse() helper to send a response,
		// rather than the failedValidationResponse() helper that we'd normally
		// use.
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Retrieve the details of the user associated with the authentication
		// token, again calling the invalidAuthenticationTokenResponse() helper
		// if no matching record was found. IMPORTANT: Notice that we are using
		// ScopeAuthentication as the first parameter here.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Call the contextSetUser() helper to add the user information to the
		// request context.
		r = app.contextSetUser(r, user)

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// requireAuthenticatedUser checks that a user is not anonymous.
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetUser() helper to retrieve the user information from
		// the request context.
		user := app.contextGetUser(r)

		// If the user is anonymous, then call the helper for sending error
		// message to inform the client that they should authenticate before
		// trying again.
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requireActivatedUser carry out these kinds of authorization checks:
// endpoints can only be accessed by users who are authenticated (not
// anonymous), and who have activated their account.
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	// Notice here that this middleware has a slightly different signature to
	// the other middleware. Instead of accepting and returning a
	// `http.Handler`, it accepts and returns a `http.HandlerFunc`.
	// This is a small change, but it makes it possible to wrap our
	// `/v1/movie**` handler functions directly with this middleware, without
	// needing to make any further conversions.

	// Rather than returning this http.HandlerFunc we assign it to the variable
	// fn.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the User struct from the request context and then check the
		// Activated field to determine whether the request should continue or
		// not.

		// Use the contextGetUser() helper to retrieve the user information from
		// the request context.
		user := app.contextGetUser(r)

		// If the user is not activated, use the inactiveAccountResponse()
		// helper to inform them that they need to activate their account.
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})

	// Wrap fn with the requireAuthenticatedUser() middleware before returning
	// it.
	// The way that we’ve set this up, this middleware now automatically calls
	// the requireAuthenticatedUser() middleware before being executed itself.
	return app.requireAuthenticatedUser(fn)
}

// requirePermission accepts a specific permission code like "movies:read" as an
// argument. It checks to see if the user permissions contains the specific
// permission code needed. If it doesn't, we should send the client a "403
// Forbidden" response.
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	// Note that the first parameter for this middleware function is the
	// permission code that we require the user to have.

	fn := func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the user from the request context.
		user := app.contextGetUser(r)

		// Get the slice of permissions for the user.
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Check if the slice includes the required permission. If it doesn't,
		// then return a 403 Forbidden response.
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// Otherwise they have the required permission so we call the next
		// handler in the chain.
		next.ServeHTTP(w, r)
	}

	// Wrap this with the requireActivatedUser() middleware before returning it.
	//
	// It means that when we use the requirePermission() middleware we'll
	// actually be carrying out three checks which together ensure that the
	// request is from an authenticated (non-anonymous), activated user, who has
	// a specific permission.
	return app.requireActivatedUser(fn)
}

// enableCORS check if the value of the request Origin header is an exact,
// case-sensitive, match for one of our trusted origins. If there is a match,
// then we should set an "Access-Control-Allow-Origin" response header which
// reflects (or echoes) back the value of the request's Origin header.
// Otherwise, we should allow the request to proceed as normal without setting
// that response header.
// It also intercepts and responds to any preflight requests.
// The purpose of this preflight request is to determine whether the real
// cross-origin request will be permitted or not.
func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A side effect of this middleware logic is that the response will be
		// different depending on the origin that the request is coming from.
		// Specifically, the value of the "Access-Control-Allow-Origin" header
		// may be different in the response, or it may not even be included at
		// all.
		// So because of this we should make sure to always set a "Vary: Origin"
		// response header to warn any caches that the response may be
		// different.
		w.Header().Add("Vary", "Origin")

		// Set a header on all responses, as the response will be different
		// depending on whether or not this header exists in the request.
		w.Header().Add("Vary", "Access-Control-Request-Method")

		// Get the value of the request's Origin header.
		origin := r.Header.Get("Origin")

		// Only run this if there's an Origin request header present AND at
		// least one trusted origin is configured.
		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			// Loop through the list of trusted origins, checking to see if the
			// request origin exactly matches one of them.
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					// If there is a match, then set a
					// 'Access-Control-Allow-Origin' response header with the
					// request origin as the value.
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// **** Identify it's a preflight cross-origin request ****
					// Check if the request has the HTTP method OPTIONS and
					// contains the "Access-Control-Request-Method" header. If
					// it does, then we treat it as a preflight request.
					if r.Method == http.MethodOptions &&
						r.Header.Get("Access-Control-Request-Method") != "" {
						// ***** Response with some special headers *****
						// Set the necessary preflight response headers to let
						// the browser know whether or not it's OK for the real
						// request to proceed.
						w.Header().Set("Access-Control-Allow-Methods",
							"OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers",
							"Authorization, Content-Type")

						// Write the headers along with a 200 OK status and
						// return from the middleware with no further action.
						w.WriteHeader(http.StatusOK)
						return
					}
				}
			}
		}

		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// metrics records custom request-level metrics for our application.
func (app *application) metrics(next http.Handler) http.Handler {
	// Initialize the new expvar variables when the middleware chain is first
	// built.
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	totalResponseSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	// The following code will be run for every request.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the Add() method to increment the number of requests received by
		// 1.
		totalRequestsReceived.Add(1)

		// ***** Record HTTP status codes *****
		// The tricky part of doing this is finding out what HTTP status code a
		// response has. Unfortunately Go doesn't make this easy — there is no
		// built-in way to examine a http.ResponseWriter to see what status code
		// is going to be sent to a client.
		//
		// The de-facto workaround is to create your own custom implementation
		// of http.ResponseWriter which records a copy of the HTTP status code
		// for future access. But doing this can be quite brittle and awkward —
		// there are several edge cases that you need to be wary of, and it can
		// cause problems if you are using any of the "additional"
		// ResponseWriter interfaces such as http.Flusher and http.Hijacker.
		//
		// Rather than making your own custom http.ResponseWriter
		// implementation, I highly recommend using the third-party httpsnoop
		// package. It's small and focused, with no additional dependencies, and
		// it makes it very easy to record the HTTP status code and size of each
		// response, along with the total processing time for each request.
		//
		// Call the httpsnoop.CaptureMetrics() function, passing in the next
		// handler in the chain along with the existing http.ResponseWriter and
		// http.Request. This returns the metrics struct.
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// On the way back up the middleware chain, increment the number of
		// responses sent by 1.
		totalResponsesSent.Add(1)

		// Get the request processing time in microseconds from httpsnoop and
		// increment the cumulative processing time.
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		// Use the Add() method to increment the count for the given status code
		// by 1. Note that the expvar map is string-keyed, so we need to use the
		// strconv.Itoa() function to convert the status code (which is an
		// integer) to a string.
		totalResponseSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
