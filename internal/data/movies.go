package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use the QueryRow() method to execute the SQL query on our connection
	// pool, passing in the args slice as a variadic parameter and scanning the
	// system-generated id, created_at and version values into the movie
	// struct.
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
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

	// Use the context.WithTimeout() function to create a context.Context which
	// carries a 3-second timeout deadline. Note that we're using the empty
	// context.Background() as the 'parent' context.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// Importantly, use defer to make sure that we cancel the context before the
	// Get() method returns.
	defer cancel()

	// Use the QueryRowContext() method to execute the query, passing in the
	// context with the deadline as the first argument.
	//
	// Then scan the response data into the fields of the Movie struct.
	// Importantly, notice that we need to convert the scan target for the
	// genres column using the pq.Array() adapter function again.
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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
        WHERE id = $5 AND version = $6
        RETURNING version`

	// Create an args slice containing the values for the placeholder
	// parameters.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query. If no matching row could be found, we know the
	// movie version has changed (or the record has been deleted) and we return
	// our custom ErrEditConflict error.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Delete deletes a specific record from the movies table.
func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1.
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `
		DELETE FROM movies
		WHERE id = $1`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query using the Exec() method, passing in the id variable
	// as the value for the placeholder parameter. The Exec() method returns a
	// sql.Result object.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number
	// of rows affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected, we know that the movies table didn't contain a
	// record with the provided ID at the moment we tried to delete it. In that
	// case we return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetAll method returns a slice of movies and pagination metadata. We've set
// this up to accept the various filter parameters as arguments.
func (m MovieModel) GetAll(title string, genres []string,
	filters Filters) ([]*Movie, Metadata, error) {
	// Construct the SQL query to retrieve all movie records.
	// Use full-text search for the title filter.
	// The window function counts the total (filtered) records.
	// Notice that we also include a secondary sort on the movie ID to ensure a
	// consistent ordering.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
		FROM movies
        WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// As our SQL query now has quite a few placeholder parameters, let's
	// collect the values for the placeholders in a slice. Notice here how we
	// call the limit() and offset() methods on the Filters struct to get the
	// appropriate values for the LIMIT and OFFSET clauses.
	args := []interface{}{title, pq.Array(genres), filters.limit(),
		filters.offset()}

	// Use QueryContext() to execute the query. This returns a sql.Rows
	// resultset containing the result.
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the resultset is
	// closed before GetAll() returns.
	defer rows.Close()

	var totalRecords int
	// Initialize an empty slice to hold the movie data.
	movies := []*Movie{}

	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual
		// movie.
		var movie Movie

		// Scan the values from the row into the Movie struct. Again, note that
		// we're using the pq.Array() adapter on the genres field here.
		err := rows.Scan(
			&totalRecords, // scan the count from the window function into totalRecords.
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Add the Movie struct to the slice.
		movies = append(movies, &movie)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any
	// error that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total record count and
	// pagination parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// If everything went OK, then return the slice of movies and pagination
	// metadata.
	return movies, metadata, nil
}

// Mocking models

var mockMovie = &Movie{
	ID:        1,
	Title:     "Casablanca",
	Year:      1960,
	Runtime:   Runtime(120),
	Genres:    []string{"drama", "documentary"},
	CreatedAt: time.Now(),
	Version:   1,
}

type MockMovieModel struct{}

// Insert inserts a new movie record. Note that this movie must not be the same
// as the mocMovie.
func (m MockMovieModel) Insert(movie *Movie) error {
	movie.ID = 2
	movie.CreatedAt = time.Now()
	movie.Version = 1

	return nil
}

// Get gets the mockMovie.
func (m MockMovieModel) Get(id int64) (*Movie, error) {
	switch id {
	case 1:
		return mockMovie, nil
	default:
		return nil, ErrRecordNotFound
	}
}

// Update updates the mockMovie.
func (m MockMovieModel) Update(movie *Movie) error {
	movie.Version = movie.Version + 1

	return nil
}

// Delete deletes the existing mockMovie.
func (m MockMovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	switch id {
	case mockMovie.ID:
		return nil
	default:
		return ErrRecordNotFound
	}
}

// GetAll filters and returns a slice of movies and pagination metadata.
func (m MockMovieModel) GetAll(title string, genres []string,
	filters Filters) ([]*Movie, Metadata, error) {
	if title != mockMovie.Title {
		return nil, Metadata{}, sql.ErrNoRows
	}

	return []*Movie{mockMovie}, Metadata{1, 10, 1, 1, 2}, nil
}
