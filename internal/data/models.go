package data

import (
	"database/sql"
	"errors"
	"time"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get()
// method when looking up a movie that doesn't exist in our database.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Create a Models struct which wraps the MovieModel. We'll add other models to
// this, like a UserModel and PermissionModel, as our build progresses.
type Models struct {
	// Set the Movies field to be an interface containing the methods that both
	// the 'real' model and mock model need to support.
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error)
	}
	Users interface {
		Insert(user *User) error
		GetByEmail(email string) (*User, error)
		Update(user *User) error
		GetForToken(tokenScope, tokenPlaintext string) (*User, error)
	}
	Tokens interface {
		New(userID int64, ttl time.Duration, scope string) (*Token, error)
		Insert(token *Token) error
		DeleteAllForUser(scope string, userID int64) error
	}
	Permissions interface {
		GetAllForUser(userID int64) (Permissions, error)
		AddForUser(userID int64, codes ...string) error
	}
}

// For ease of use, we also add a New() method which returns a Models struct
// containing the initialized MovieModel and initialized UserModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
	}
}

// Create a helper function which returns a Models instance containing the mock
// models only.
func NewMockModels() Models {
	return Models{
		Movies:      MockMovieModel{},
		Users:       MockUserModel{},
		Tokens:      MockTokenModel{},
		Permissions: MockPermissionModel{},
	}
}
