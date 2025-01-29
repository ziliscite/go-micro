package main

import (
	"net/http"
)

type response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *application) Broker(w http.ResponseWriter, r *http.Request) {
	payload := response{
		Error:   false,
		Message: "Hit the broker",
		Data:    nil,
	}

	err := app.write(w, http.StatusOK, payload)
	if err != nil {
		app.error(w, http.StatusInternalServerError, err)
	}
}
