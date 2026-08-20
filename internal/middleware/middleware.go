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
	jwtUserIDValueName string
	jwt                jwebtoken
	headerName         string
}

func NewMiddleware(hName string, jwt jwebtoken, jwtUserIDValueName string) *middleware {
	return &middleware{
		jwt:                jwt,
		headerName:         hName,
		jwtUserIDValueName: jwtUserIDValueName,
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

		ctx := context.WithValue(r.Context(), m.jwtUserIDValueName, claims[m.jwtUserIDValueName])

		return next(w, r.WithContext(ctx))
	}
}
