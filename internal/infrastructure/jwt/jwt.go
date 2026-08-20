package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	errs "github.com/newaccg/todo-api/internal/errors"
)

type jwebtoken struct {
	expirationTime     time.Duration
	secret             string
	jwtUserIDValueName string
}

func NewJWT(JWTExpirationTime time.Duration, JWTSecret, jwtUserIDValueName string) *jwebtoken {
	return &jwebtoken{
		expirationTime:     JWTExpirationTime,
		secret:             JWTSecret,
		jwtUserIDValueName: jwtUserIDValueName,
	}
}

func (j *jwebtoken) GenerateJWT(userId int64) (string, error) {
	claims := jwt.MapClaims{
		"exp":                time.Now().Add(j.expirationTime).Unix(),
		j.jwtUserIDValueName: userId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *jwebtoken) ValidateAndGetClaimsFromJWT(token string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ErrInvalidSignMethod
		}

		return []byte(j.secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errs.ErrTokenExpired
		}
		return nil, errs.ErrInvalidToken
	}

	if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return nil, errs.ErrTokenExpired
		}

		return claims, nil
	}

	return nil, errs.ErrInvalidToken
}
