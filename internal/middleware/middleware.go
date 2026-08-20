package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"

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

	rateLimitRefillPerSecond int
	rateLimitRefreshDuration time.Duration
	visitors                 map[string]*visitor
	mu                       sync.Mutex
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewMiddleware(hName string, jwt jwebtoken, jwtUserIDValueName string, refillSpeed int, refreshDur time.Duration) *middleware {
	res := &middleware{
		jwt:                jwt,
		headerName:         hName,
		jwtUserIDValueName: jwtUserIDValueName,

		rateLimitRefillPerSecond: refillSpeed,
		rateLimitRefreshDuration: refreshDur,
		visitors:                 make(map[string]*visitor),
	}

	go res.cleanVisitors()

	return res
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

func (m *middleware) RateLimit(next handler.CustomHandler, bucketSize int) handler.CustomHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return err
		}

		limiter := m.getVisitor(ip, r.URL.Path, bucketSize)
		if !limiter.Allow() {
			return errs.ErrTooManyRequests
		}

		return next(w, r)
	}
}

func (m *middleware) getVisitor(ip, path string, bucketSize int) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s:%s", ip, path)
	v, ok := m.visitors[key]
	if !ok {
		limiter := rate.NewLimiter(rate.Limit(m.rateLimitRefillPerSecond), bucketSize)
		m.visitors[key] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}

		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (m *middleware) cleanVisitors() {
	for {
		time.Sleep(time.Minute)

		m.mu.Lock()
		for ip, v := range m.visitors {
			if time.Since(v.lastSeen) > m.rateLimitRefreshDuration {
				delete(m.visitors, ip)
			}
		}
		m.mu.Unlock()
	}
}
