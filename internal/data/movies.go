package data

import (
	"time"

	"github.com/cedrickchee/skel/internal/validator"
)

// Annotate the Movie struct with struct tags to control how the keys appear in
// the JSON-encoded output.
type Movie struct {
	ID        int64     `json:"id"`             // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`              // Timestamp for when the movie is added to our database
	Title     string    `json:"title"`          // Movie title
	Year      int32     `json:"year,omitempty"` // Movie release year
	// Use the Runtime type instead of int32. Note that the omitempty directive
	// will still work on this: if the Runtime field has the underlying value 0,
	// then it will be considered empty and omitted -- and the MarshalJSON()
	// method we just made won't be called at all.
	Runtime Runtime  `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres  []string `json:"genres,omitempty"`  // Slice of genres for the movie (romance, comedy, etc.)
	Version int32    `json:"version"`           // The version number starts at 1 and will be incremented each time the movie information is updated
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// Use the Check() method to execute our validation checks. This will add
	// the provided key and error message to the errors map if the check does
	// not evaluate to true. For example, in the first line here we "check that
	// the title is not equal to the empty string". In the second, we "check
	// that the length of the title is less than or equal to 500 bytes" and so
	// on.
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	// Note that we're using the Unique helper in the line below to check that
	// all values in the movie.Genres slice are unique.
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
