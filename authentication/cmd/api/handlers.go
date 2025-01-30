package main

import (
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ziliscite/go-micro-authentication/internal/data"
	"net/http"
)

type response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	var request struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	err := app.readBody(w, r, &request)
	if err != nil {
		app.error(w, http.StatusBadRequest, err)
		return
	}

	user := data.User{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
	}

	err = user.Password.Set(request.Password)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.repo.Insert(&user)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			app.error(w, http.StatusConflict, errors.New("a user with this email address already exists"))
		default:
			app.serverError(w, err)
		}
		return
	}

	if err = app.write(w, http.StatusAccepted, response{
		Error:   false,
		Message: "User Created",
		Data:    user,
	}); err != nil {
		app.serverError(w, err)
	}
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readBody(w, r, &request)
	if err != nil {
		app.error(w, http.StatusBadRequest, err)
		return
	}

	user, err := app.repo.GetByEmail(request.Email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.invalidCredentials(w)
		default:
			app.serverError(w, err)
		}
		return
	}

	valid, err := user.Password.PasswordMatches(request.Password)
	if err != nil || !valid {
		app.invalidCredentials(w)
		return
	}

	if err = app.write(w, http.StatusAccepted, response{
		Error:   false,
		Message: "Authenticated",
		Data:    user,
	}); err != nil {
		app.serverError(w, err)
	}
}
