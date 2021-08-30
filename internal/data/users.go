package data

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User struct represents an individual user. Importantly, notice how we are
// using the json:'-' struct tag to prevent the Password and Version fields
// appearing in any output when we encode it to JSON. Also notice that the
// Password field uses the custom password type defined below.
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// A custom password type which is a struct containing the plaintext and hashed
// versions of the password for a user. The plaintext field is a *pointer* to a
// string, so that we're able to distinguish between a plaintext password not
// being present in the struct at all, versus a plaintext password which is the
// empty string "".
type password struct {
	plaintext *string
	hash      []byte
}

// Set method calculates the bcrypt hash of a plaintext password, and stores
// both the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	// Cost parameter: we use a cost of 12.
	// The higher the cost, the slower and more computationally expensive it is
	// to generate the hash. There is a balance to be struck here — we want the
	// cost to be prohibitively expensive for attackers, but also not so slow
	// that it harms the user experience of our API.
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

// Matches method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
