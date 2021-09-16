package data

import (
	"reflect"
	"testing"
	"time"
)

func TestUserModelGetByEmail(t *testing.T) {
	// Skip the test if the `-short` flag is provided when running the test.
	if testing.Short() {
		t.Skip("postgresql: skipping integration test")
	}

	// Set up a suite of table-driven tests and expected results.
	tz, err := time.LoadLocation("Asia/Singapore")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name      string
		userID    int
		email     string
		wantUser  *User
		wantError error
	}{
		{
			name:   "Valid Email",
			userID: 1,
			email:  "alice@example.com",
			wantUser: &User{
				ID:        1,
				Name:      "Alice Jones",
				Email:     "alice@example.com",
				CreatedAt: time.Date(2021, 9, 17, 1, 10, 0, 0, tz),
				Activated: true,
				Password:  password{hash: []byte("013d7d16d7ad4fefb61bd95b765c8ceb")},
				Version:   1,
			},
			wantError: nil,
		},
		{
			name:      "Non-existent Email",
			email:     "john@example.com",
			userID:    2,
			wantUser:  nil,
			wantError: ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize a connection pool to our test database, and defer a
			// call to the teardown function, so it is always run immediately
			// before this sub-test returns.
			db, teardown := newTestDB(t)
			defer teardown()

			// Create a new instance of the UserModel.
			m := UserModel{db}

			// Call the UserModel.GetByEmail() method and check that the return
			// value and error match the expected values for the sub-test.
			user, err := m.GetByEmail(tt.email)

			if err != tt.wantError {
				t.Errorf("want %v; got %s", tt.wantError, err)
			}

			if !reflect.DeepEqual(user, tt.wantUser) {
				t.Errorf("want %+v; got %+v", tt.wantUser, user)
			}
		})
	}
}

/*
Run:

$ go test -v -run ^TestUserModelGetByEmail$ github.com/cedrickchee/skel/internal/data
*/

/*
Skipping long-running tests

When your tests take a long time, you might decide that you want to skip
specific long-running tests under certain circumstances. For example, you might
decide to only run your integration tests before committing a change, instead of
more frequently during development.

A common and idiomatic way to skip long-running tests is to use the
[testing.Short()](https://golang.org/pkg/testing/#Short) function to check for
the presence of a `-short` flag, like we have in the code above.

When we run our tests with the `-short` flag, then our `TestUserModelGetByEmail`
test will be skipped.

$ go test -v -short ./...
*/
