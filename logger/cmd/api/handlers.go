package main

import (
	"github.com/ziliscite/go-micro-logger/internal/data"
	"github.com/ziliscite/go-micro-logger/internal/repository"

	"context"
	"errors"
	"net/http"
	"time"
)

func (app *application) writeLog(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	err := app.readBody(w, r, &request)
	if err != nil {
		app.error(w, http.StatusBadRequest, err)
		return
	}

	entry := data.Entry{
		Title:   request.Title,
		Content: request.Content,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err = app.repo.Insert(ctx, &entry)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrDuplicateEntry):
			app.error(w, http.StatusConflict, err)
		case errors.Is(err, repository.ErrInvalidData):
			app.error(w, http.StatusBadRequest, err)
		case errors.Is(err, repository.ErrDatabaseTimeout):
			app.error(w, http.StatusGatewayTimeout, err)
		default:
			app.serverError(w, err)
		}
		return
	}

	if err = app.write(w, http.StatusAccepted, response{
		Error:   false,
		Message: "Log Inserted",
		Data:    entry,
	}); err != nil {
		app.serverError(w, err)
	}
}

func (app *application) listLogs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	entries, err := app.repo.GetAll(ctx)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			app.error(w, http.StatusNotFound, err)
		case errors.Is(err, repository.ErrInvalidID):
			app.error(w, http.StatusBadRequest, err)
		default:
			app.serverError(w, err)
		}
		return
	}

	if err = app.write(w, http.StatusOK, response{
		Error:   false,
		Message: "Logs Fetched",
		Data:    entries,
	}); err != nil {
		app.serverError(w, err)
	}
}
