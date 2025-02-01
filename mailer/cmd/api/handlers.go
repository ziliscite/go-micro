package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func (app *application) send(w http.ResponseWriter, r *http.Request) {
	var request struct {
		From    string `json:"from,omitempty"`
		To      string `json:"to,omitempty"`
		Subject string `json:"subject,omitempty"`
		Message string `json:"message,omitempty"`
	}

	err := app.readBody(w, r, &request)
	if err != nil {
		app.error(w, http.StatusBadRequest, err)
		return
	}

	go func() {
		// In case of panic / unhandled error
		defer func() {
			if err := recover(); err != nil {
				slog.Error(fmt.Sprintf("%v", err))
			}
		}()

		// Send email in a goroutine
		if err = app.mailer.SendMessage(Message{
			Data:    request.Message,
			From:    request.From,
			To:      request.To,
			Subject: request.Subject,
		}); err != nil {
			slog.Error(err.Error())
		}
	}()

	if err = app.write(w, http.StatusAccepted, response{
		Error:   false,
		Message: fmt.Sprintf("Email is being sent to %s, please wait.", request.To),
	}); err != nil {
		app.error(w, http.StatusInternalServerError, err)
	}
}
