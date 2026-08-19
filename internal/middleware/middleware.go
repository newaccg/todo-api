package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/handler"
)

type jwebtoken interface {
	ValidateAndGetClaimsFromJWT(token string) (jwt.MapClaims, error)
}

type middleware struct {
	jwt        jwebtoken
	headerName string
}

func NewMiddleware(hName string, jwt jwebtoken) *middleware {
	return &middleware{
		jwt:        jwt,
		headerName: hName,
	}
}

func (m *middleware) Auth(next handler.CustomHandler) handler.CustomHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		token := r.Header.Get(m.headerName)
		if token == "" {
			return errs.ErrEmptyToken
		}

		claims, err := m.jwt.ValidateAndGetClaimsFromJWT(token)
		if err != nil {
			return fmt.Errorf("could not process token \"%s\": %w", token, err)
		}

		ctx := context.WithValue(r.Context(), "userID", claims["userID"])

		return next(w, r.WithContext(ctx))
	}
}
