package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	errs "github.com/newaccg/todo-api/internal/errors"
)

type CustomHandler func(w http.ResponseWriter, r *http.Request) error

func Handle(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info(
			"incoming HTTP request",
			"address", r.RemoteAddr,
			"path", r.URL.Path,
			"method", r.Method,
		)
		err := f(w, r)

		if err != nil {
			logMessage := "API error"

			var resp struct {
				Err  string `json:"error"`
				Code int    `json:"code"`
			}

			switch {
			case errors.Is(err, errs.ErrInvalidID):
				resp.Err = "invalid ID"
				resp.Code = http.StatusBadRequest

			case errors.Is(err, errs.ErrInvalidJSON):
				resp.Err = "invalid JSON"
				resp.Code = http.StatusBadRequest

			case errors.Is(err, errs.ErrInvalidURLValue):
				resp.Err = "invalid URl value(s)"
				resp.Code = http.StatusBadRequest

			case errors.Is(err, errs.ErrUserNotFound):
				resp.Err = "user with given email doesn't exist"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrWrongPassword):
				resp.Err = "wrong password"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrTaskNotFound):
				resp.Err = "task not found"
				resp.Code = http.StatusNotFound

			case errors.Is(err, errs.ErrInvalidSignMethod):
				resp.Err = "invalid signing method"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrTokenExpired):
				resp.Err = "token expired"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrInvalidSignMethod):
				resp.Err = "invalid signing method"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrInvalidToken):
				resp.Err = "invalid token"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrEmailExists):
				resp.Err = "user with given email already exists"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrEmptyToken):
				resp.Err = "given token is empty"
				resp.Code = http.StatusUnauthorized

			case errors.Is(err, errs.ErrEmptyField):
				resp.Err = "given field(s) is(are) empty"
				resp.Code = http.StatusBadRequest

			case errors.Is(err, errs.ErrTooManyRequests):
				resp.Err = "too many requests"
				resp.Code = http.StatusTooManyRequests

			default:
				resp.Err = "internal server error"
				resp.Code = http.StatusInternalServerError
				logMessage = "internal error"
			}

			slog.Error(
				logMessage,
				"error", err,
			)

			writeJSONWithCode(w, resp, resp.Code)
		}
	}
}

func writeJSONWithCode(w http.ResponseWriter, msg any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		slog.Error(
			"could not encode JSON",
			"error", err,
		)
	}
}
