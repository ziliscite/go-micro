package data

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is the structure which holds one user from the database.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Password  password  `json:"-"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) Hashed() []byte {
	return u.Password.hash
}

func (u *User) SetHashed(hashedPassword []byte) {
	u.Password.hash = hashedPassword
}

type password struct {
	plaintext string
	hash      []byte
}

// PasswordMatches uses Go's bcrypt package to compare a user-supplied password
// with the hash we have stored for a given user in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (p *password) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// Set method calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = plaintextPassword
	p.hash = hash

	return nil
}
