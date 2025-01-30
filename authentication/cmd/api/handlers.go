package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ziliscite/go-micro-authentication/internal/data"
	"github.com/ziliscite/go-micro-authentication/internal/repository"
	"log/slog"
	"net/http"
)

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

	ctx, cancel := context.WithTimeout(r.Context(), repository.DBTimeout)
	defer cancel()

	err = app.repo.Insert(ctx, &user)
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

	ctx, cancel := context.WithTimeout(r.Context(), repository.DBTimeout)
	defer cancel()

	user, err := app.repo.GetByEmail(ctx, request.Email)
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

	// I assume this will be changed to some pub sub and call grpc stuff
	err = app.log("Authenticated", fmt.Sprintf("%s authenticated successfully", user.Email))
	if err != nil {
		// don't do anything. why would we want to make it error here?
		// user is authenticated, a missing log aint doin shit
		slog.Error("Failed to log authentication", "error", err)
	}

	if err = app.write(w, http.StatusAccepted, response{
		Error:   false,
		Message: "Authenticated",
		Data:    user,
	}); err != nil {
		app.serverError(w, err)
	}
}

func (app *application) log(title, content string) error {
	var entry struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	entry.Title = title
	entry.Content = content

	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://logger/v1/logs", bytes.NewBuffer(payload))
	// url is composed of [hostname]:[port]/[service name in the docker image]/[method]
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
