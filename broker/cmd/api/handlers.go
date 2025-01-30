package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

type request struct {
	Action string `json:"action"`
	Auth   auth   `json:"auth,omitempty"`
}

type auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) gateway(w http.ResponseWriter, r *http.Request) {
	var req request

	err := app.readBody(w, r, &req)
	if err != nil {
		app.error(w, http.StatusBadRequest, err)
		return
	}

	// Match the request action
	switch req.Action {
	case "authenticate":
		app.authenticate(w, req.Auth)
	default:
		app.error(w, http.StatusNotImplemented, errors.New("unknown action"))
	}
}

func (app *application) authenticate(w http.ResponseWriter, a auth) {
	// Create the payload
	payload, err := json.Marshal(a)
	if err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}

	// Call the authentication microservice
	req, err := http.NewRequest(http.MethodPost, "http://authentication/v1/authenticate", bytes.NewBuffer(payload))
	// url is composed of [hostname]:[port]/[service name in the docker image]/[method]
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}
	defer resp.Body.Close()

	var message string
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		message = "invalid credentials"
	default:
		message = "auth service could not process your request"
	}

	// Check the status code
	if resp.StatusCode != http.StatusAccepted {
		app.error(w, resp.StatusCode, errors.New(message))
		return
	}

	// Decode the response
	var jsonResp response
	if err = json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}

	if jsonResp.Error {
		app.error(w, http.StatusUnauthorized, errors.New(jsonResp.Message))
		return
	}

	// Add the auth token to the response
	if err = app.write(w, http.StatusOK, jsonResp); err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}
}
