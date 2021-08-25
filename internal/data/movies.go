package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/cedrickchee/skel/internal/validator"
	"github.com/lib/pq"
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

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type MovieModel struct {
	DB *sql.DB
}

// The Insert() method accepts a pointer to a movie struct, which should contain
// the data for the new record.
func (m MovieModel) Insert(movie *Movie) error {
	// Define the SQL query for inserting a new record in the movies table and
	// returning the system-generated data.
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// Create an args slice containing the values for the placeholder parameters
	// from the movie struct. Declaring this slice immediately next to our SQL
	// query helps to make it nice and clear *what values are being used where*
	// in the query.
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Use the QueryRow() method to execute the SQL query on our connection
	// pool, passing in the args slice as a variadic parameter and scanning the
	// system-generated id, created_at and version values into the movie
	// struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get fetches a specific record from the movies table.
func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts
	// auto-incrementing at 1 by default, so we know that no movies will have ID
	// values less than that. To avoid making an unnecessary database call, we
	// take a shortcut and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Define the SQL query for retrieving the movie data.
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`

	// Declare a Movie struct to hold the data returned by the query.
	var movie Movie

	// Execute the query using the QueryRow() method, passing in the provided id
	// value as a placeholder parameter, and scan the response data into the
	// fields of the Movie struct. Importantly, notice that we need to convert
	// the scan target for the genres column using the pq.Array() adapter
	// function again.
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handle any errors. If there was no matching movie found, Scan() will
	// return a sql.ErrNoRows error. We check for this and return our custom
	// ErrRecordNotFound error instead.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Otherwise, return a pointer to the Movie struct.
	return &movie, nil
}

// Update updates a specific record in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	// Declare the SQL query for updating the record and returning the new
	// version number.
	query := `
        UPDATE movies
        SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
        WHERE id = $5
        RETURNING version`

	// Create an args slice containing the values for the placeholder
	// parameters.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}

	// Use the QueryRow() method to execute the query, passing in the args slice
	// as a variadic parameter and scanning the new version value into the movie
	// struct.
	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

// Add a placeholder method for deleting a specific record from the movies
// table.
func (m MovieModel) Delete(id int64) error {
	return nil
}

// Mocking models

var mockMovie = &Movie{
	ID:      1,
	Title:   "Casablance",
	Year:    1960,
	Runtime: Runtime(120),
	Genres:  []string{"drama", "documentary"},
}

type MockMovieModel struct{}

func (m MockMovieModel) Insert(movie *Movie) error {
	return nil
}

func (m MockMovieModel) Get(id int64) (*Movie, error) {
	switch id {
	case 1:
		return mockMovie, nil
	default:
		return nil, ErrRecordNotFound
	}
}

func (m MockMovieModel) Update(movie *Movie) error {
	return nil
}

func (m MockMovieModel) Delete(id int64) error {
	return nil
}
