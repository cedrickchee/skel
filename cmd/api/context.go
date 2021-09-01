package main

import (
	"context"
	"net/http"

	"github.com/cedrickchee/skel/internal/data"
)

// Define a custom type for the request context keys, with the underlying type
// string. This helps prevent naming collisions between our code and any
// third-party packages which are also using the request context to store
// information.
type contextKey string

// Convert the string "user" to a contextKey type and assign it to the
// userContextKey constant. We'll use this constant as the key for getting and
// setting user information in the request context.
const userContextKey = contextKey("user")

// contextSetUser method returns a new copy of the request with the provided
// User struct added to the context. Note that we use our userContextKey
// constant as the key.
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGetUser method retrieves the User struct from the request context. The
// only time that we'll use this helper is when we logically expect there to be
// User struct value in the context, and if it doesn't exist it will firmly be
// an 'unexpected' error. As we discussed earlier in the book, it's OK to panic
// in those circumstances.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
