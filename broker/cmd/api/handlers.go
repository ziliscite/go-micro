package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ziliscite/go-micro-broker/event"
	logs "github.com/ziliscite/go-micro-broker/proto/genproto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"net/rpc"
)

func (app *application) broker(w http.ResponseWriter, r *http.Request) {
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
	Log    log    `json:"log,omitempty"`
	Mail   mail   `json:"mail,omitempty"`
}

type auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type log struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type mail struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Subject string `json:"subject,omitempty"`
	Message string `json:"message,omitempty"`
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
	case "log":
		//app.log(w, req.Log) -- http
		//app.pushLog(w, req.Log) -- message broker
		app.logRPC(w, req.Log)

	case "mail":
		app.sendMail(w, req.Mail)
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

func (app *application) log(w http.ResponseWriter, l log) {
	// Create the payload
	payload, err := json.Marshal(l)
	if err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}

	// Call the authentication microservice
	req, err := http.NewRequest(http.MethodPost, "http://logger/v1/logs", bytes.NewBuffer(payload))
	// url is composed of [hostname]:[port]/[service name in the docker image]/[method]
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}
	defer resp.Body.Close()

	var message string
	switch resp.StatusCode {
	case http.StatusConflict:
		message = "a conflict occurred"
	case http.StatusBadRequest:
		message = "invalid log data"
	case http.StatusGatewayTimeout:
		message = "gateway timeout"
	case http.StatusNotFound:
		message = "resource not found"
	default:
		message = "log service could not process your request"
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
		app.error(w, http.StatusInternalServerError, errors.New(jsonResp.Message))
		return
	}

	if err = app.write(w, http.StatusOK, jsonResp); err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) sendMail(w http.ResponseWriter, m mail) {
	// Create the payload
	payload, err := json.Marshal(m)
	if err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, "http://mailer/v1/send", bytes.NewBuffer(payload))
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}
	defer resp.Body.Close()

	var message string
	switch resp.StatusCode {
	case http.StatusBadRequest:
		message = "invalid email data"
	default:
		message = "mailer could not process your request"
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
		app.error(w, http.StatusInternalServerError, errors.New(jsonResp.Message))
		return
	}

	// Add the auth token to the response
	if err = app.write(w, http.StatusOK, jsonResp); err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) pushLog(w http.ResponseWriter, l log) {
	err := app.pushToQueue(l.Title, l.Content)
	if err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}

	if err = app.write(w, http.StatusAccepted, response{
		Error:   false,
		Message: "Log pushed to queue",
	}); err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}
}

// same pattern to publish shit to queue
func (app *application) pushToQueue(name, msg string) error {
	pub, err := event.NewPublisher(app.rabbit)
	if err != nil {
		return err
	}

	p := log{
		Title:   name,
		Content: msg,
	}

	pj, err := json.Marshal(&p)
	if err != nil {
		return err
	}

	// might wanna break it into some log types
	return pub.Push(string(pj), "log.INFO")
}

func (app *application) logRPC(w http.ResponseWriter, l log) {
	type rpcPayload struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	// Get rpc client
	client, err := rpc.Dial("tcp", "logger:5001") // same name in docker compose
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}
	defer client.Close()

	// Create type that exactly matches the on the rpc
	payload := rpcPayload{
		Name: l.Title,
		Data: l.Content,
	}

	var res string // response from the rpc

	// The service method name must exactly match the one on the rpc
	//
	// Must start with a capital letter to be exported (so that it works)
	err = client.Call("RPCServer.LogInfo", payload, &res)
	if err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}

	if err = app.write(w, http.StatusOK, response{
		Error:   false,
		Message: res,
	}); err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) logGRPC(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	var l log

	err := app.readBody(w, r, &l)
	if err != nil {
		app.error(w, http.StatusBadRequest, err)
		return
	}

	// Get grpc client
	// need credentials, but since we using docker...
	client, err := grpc.NewClient("logger:50001", grpc.WithTransportCredentials(insecure.NewCredentials())) // same name in docker compose
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}
	defer client.Close()

	// This should (along with the client, which can be treated as database connection)
	// be injected as a dependency
	//
	// On the application / handler struct -- like a repository
	logClient := logs.NewLogServiceClient(client)

	resp, err := logClient.WriteLog(ctx, &logs.LogRequest{
		Entry: &logs.Log{
			Name: l.Title,
			Data: l.Content,
		},
	})
	if err != nil {
		app.error(w, http.StatusServiceUnavailable, err)
		return
	}

	if err = app.write(w, http.StatusOK, response{
		Error:   false,
		Message: resp.Response,
	}); err != nil {
		app.error(w, http.StatusInternalServerError, err)
		return
	}
}
