package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

type jwebtoken struct {
	accessTokenExpirationTime  time.Duration
	refreshTokenExpirationTime time.Duration

	secret             string
}

const JWTUserIDKey = "userID"

func NewJWT(accessExpirationTime, refreshExpirationTime time.Duration, JWTSecret string) *jwebtoken {
	return &jwebtoken{
		accessTokenExpirationTime:  accessExpirationTime,
		refreshTokenExpirationTime: refreshExpirationTime,
		secret:                     JWTSecret,
	}
}

func (j *jwebtoken) GenerateAccessToken(userID int64) (*model.Token, error) {
	return j.generateJWT(userID, j.accessTokenExpirationTime)
}

func (j *jwebtoken) GenerateRefreshToken(userID int64) (*model.Token, error) {
	return j.generateJWT(userID, j.refreshTokenExpirationTime)
}

func (j *jwebtoken) ValidateAndGetClaimsFromJWT(token string) (*model.Claims, error) {
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
		customClaims := &model.Claims{}

		customClaims.UserID, err = claimValueToInt64(claims[JWTUserIDKey])
		if err != nil {
			return nil, err
		}

		customClaims.ExpirationTime, err = claimValueToInt64(claims["exp"])
		if err != nil {
			return nil, err
		}

		jti := claims["jti"].(string)
		customClaims.TokenID = jti

		if time.Now().Unix() > customClaims.ExpirationTime {
			return nil, errs.ErrTokenExpired
		}

		return customClaims, nil
	}

	return nil, errs.ErrInvalidToken
}

func (j *jwebtoken) generateJWT(userID int64, dur time.Duration) (*model.Token, error) {
	exp := time.Now().Add(dur).UTC().Unix()

	uuid, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	jti := uuid.String()

	claims := jwt.MapClaims{
		"exp":                exp,
		"jti":                jti,
		JWTUserIDKey: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	resultToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return nil, err
	}

	resultClaims := model.Claims{
		ExpirationTime: exp,
		UserID:         userID,
		TokenID:        jti,
	}

	res := &model.Token{
		Token:  resultToken,
		Claims: resultClaims,
	}

	return res, nil
}

func claimValueToInt64(val any) (int64, error) {
	// we cannot convert directly to int64 because jwt.MapClaims converts all numeric values to float64
	// so we have to do this intermediate step
	floatVal, ok := val.(float64)
	if !ok {
		return 0, errors.New("could not convert ID from JWT float64")
	}

	// this is guaranteed to be converted from float64 to int64
	// so we don't have to check for success
	return int64(floatVal), nil
}
