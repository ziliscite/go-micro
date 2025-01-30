package repository

import (
	"context"
	"database/sql"
	"github.com/ziliscite/go-micro-authentication/internal/data"
	"log/slog"
	"time"
)

const dbTimeout = time.Second * 3

// New is the function used to create an instance of the data package. It returns the type
// Model, which embeds all the types we want to be available to our application.
func New(dbPool *sql.DB) Repository {
	return Repository{
		db: dbPool,
	}
}

// Repository is the type for this package. Note that any model that is included as a member
// in this type is available to us throughout the application, anywhere that the
// app variable is used, if the model is also added in the New function.
type Repository struct {
	db *sql.DB
}

// GetAll returns a slice of all users, sorted by last name
func (r Repository) GetAll() ([]*data.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT id, email, first_name, last_name, password, user_active, created_at, updated_at
	FROM users ORDER BY last_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*data.User

	for rows.Next() {
		var hashed []byte

		var user data.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&hashed,
			&user.Active,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			slog.Error("Unable to scan row", "error", err)
			return nil, err
		}

		user.SetHashed(hashed)

		users = append(users, &user)
	}

	return users, nil
}

// GetByEmail returns one user by email
func (r Repository) GetByEmail(email string) (*data.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
	SELECT id, email, first_name, last_name, password, user_active, created_at, updated_at 
	FROM users WHERE email = $1
	`

	var hashed []byte
	var user data.User
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&hashed,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	user.SetHashed(hashed)

	return &user, nil
}

// GetOne returns one user by id
func (r Repository) GetOne(id int) (*data.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		SELECT id, email, first_name, last_name, password, user_active, created_at, updated_at 
		FROM users WHERE id = $1
	`

	var hashed []byte

	var user data.User
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&hashed,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	user.SetHashed(hashed)

	return &user, nil
}

// Update updates one user in the database, using the information
// stored in the receiver u
func (r Repository) Update(user *data.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE users SET 
		email = $1, first_name = $2, last_name = $3,
		user_active = $4, updated_at = $5
		WHERE id = $6 AND updated_at = $7
	`

	_, err := r.db.ExecContext(ctx, stmt,
		user.Email, user.FirstName,
		user.LastName, user.Active,
		time.Now(), user.ID,
		user.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes one user from the database, by User.ID
func (r Repository) Delete(user *data.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, stmt, user.ID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteByID deletes one user from the database, by ID
func (r Repository) DeleteByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM users where id = $1`

	_, err := r.db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (r Repository) Insert(user *data.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		INSERT INTO users (email, first_name, last_name, password)
		VALUES ($1, $2, $3, $4) RETURNING id
	`

	var newID int
	if err := r.db.QueryRowContext(ctx, stmt,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Hashed(),
	).Scan(&newID); err != nil {
		return err
	}

	user.ID = newID

	return nil
}

// ResetPassword is the method we will use to change a user's password.
func (r Repository) ResetPassword(user *data.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE users SET password = $1 WHERE id = $2 AND updated_at = $3`
	_, err := r.db.ExecContext(ctx, stmt, user.Hashed(), user.ID, user.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}
