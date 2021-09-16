package data

import (
	"context"
	"database/sql"
	"io/ioutil"
	"testing"
	"time"
)

// newTestDB is a helper function which:
// - Creates a new *sql.DB connection pool for the test database.
// - Executes the setup.sql script to create the database tables and dummy data.
// - Returns an anonymous function which executes the teardown.sql script and
// closes the connection pool.
func newTestDB(t *testing.T) (*sql.DB, func()) {
	// Call the openDB() helper function to create the connection pool for the
	// test database.
	dsn := "postgres://test_skel:TESTpa55word@localhost/test_skel"
	db, err := openDB(dsn)
	if err != nil {
		t.Fatal(err)
	}

	// Read the setup SQL script from file and execute the statements.
	script, err := ioutil.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	// Return the connection pool and an anonymous function which reads and
	// executes the teardown script, and closes the connection pool. We can
	// assign this anonymous function and call it later once our test has
	// completed.
	return db, func() {
		script, err := ioutil.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}

		db.Close()
	}
}

// The openDB() function returns a sql.DB connection pool.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
