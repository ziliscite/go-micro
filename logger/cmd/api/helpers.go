package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

var (
	ErrBadlyFormattedJSON = errors.New("body contains badly-formed JSON")
	ErrInvalidTypeJSON    = errors.New("body contains incorrect JSON type")
)

func (app *application) write(w http.ResponseWriter, code int, data any, headers ...http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) error(w http.ResponseWriter, code int, e error) {
	var res response
	res.Error = true
	res.Message = e.Error()

	slog.Error("some problems occurred", "error", e.Error(), "code", code)

	err := app.write(w, code, res, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	message := "the server encountered a problem and could not process your request"
	app.error(w, http.StatusInternalServerError, errors.New(message))
}

func (app *application) readBody(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("%w (at character %d)", ErrBadlyFormattedJSON, syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return ErrBadlyFormattedJSON

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("%w for field %q", ErrInvalidTypeJSON, unmarshalTypeError.Field)
			}
			return fmt.Errorf("%w (at character %d)", ErrInvalidTypeJSON, unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key: '%s'", strings.Trim(fieldName, "\""))

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
