package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// Retrieve the 'id' URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an
// error.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// When httprouter is parsing a request, any interpolated URL parameters
	// will be stored in the request context. We can use the ParamsFromContext()
	// function to retrieve a slice containing these parameter names and values.
	params := httprouter.ParamsFromContext(r.Context())

	// We can then use the ByName() method to get the value of the 'id'
	// parameter from the slice. In our project all movies will have a unique
	// positive integer ID, but the value returned by ByName() is always a
	// string. So we try to convert it to a base 10 integer (with a bit size of
	// 64). If the parameter couldn't be converted, or is less than 1, we know
	// the ID is invalid so we return an error.
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelope map[string]interface{}

// Define a writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code to send, the data to encode to
// JSON, and a header map containing any additional HTTP headers we want to
// include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int,
	data envelope, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	// Use the json.MarshalIndent() function so that whitespace is added to the
	// encoded JSON. Here we use no line prefix ('') and tab indents ('\t') for
	// each element.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')

	// At this point, we know that we won't encounter any more errors before
	// writing the response, so it's safe to add any headers that we want to
	// include. We loop through the header map and add each header to the
	// http.ResponseWriter header map. Note that it's OK if the provided header
	// map is nil. Go doesn't throw an error if you try to range over (or
	// generally, read from) a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the 'Content-Type: application/json' header, then write the status
	// code and JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request,
	dst interface{}) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize a new json.Decoder instance which reads from the request body,
	// and call the DisallowUnknownFields() method on it before decoding. This
	// means that if the JSON from the client now includes any field which
	// cannot be mapped to the target destination, the decoder will return an
	// error instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body to the destination (could be a struct).
	err := dec.Decode(dst)
	if err != nil {
		// If there is an error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// Use the errors.As() function to check whether the error has the type
		// *json.SyntaxError. If it does, then return a plain-english error
		// message which includes the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF
		// error for syntax errors in the JSON. So we check for this using
		// errors.Is() and return a generic error message. There is an open
		// issue regarding this at https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// Likewise, catch any *json.UnmarshalTypeError errors. These occur when
		// the JSON value is the wrong type for the target destination. If the
		// error relates to a specific field, then we include that in our error
		// message to make it easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is
		// empty. We check for this with errors.Is() and return a plain-english
		// error message instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If the JSON contains a field which cannot be mapped to the target
		// destination then Decode() will now return an error message in the
		// format "json: unknown field "<name>"". We check for this, extract the
		// field name from the error, and interpolate it into our custom error
		// message. Note that there's an open issue at
		// https://github.com/golang/go/issues/29035 regarding turning this into
		// a distinct error type in the future.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// If the request body exceeds 1MB in size the decode will now fail with
		// the error "http: request body too large". There is an open issue
		// about turning this into a distinct error type at
		// https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		// A json.InvalidUnmarshalError error will be returned if we pass a non-nil
		// pointer to Decode(). We catch this and panic, rather than returning an error
		// to our handler. At the end of this chapter we'll talk about panicking
		// versus returning errors, and discuss why it's an appropriate thing to do in
		// this specific situation.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// For anything else, return the error message as-is.
		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value this
	// will return an io.EOF error. So if we get anything else, we know that
	// there is additional data in the request body and we return our own custom
	// error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
