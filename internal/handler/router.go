package handler

import (
	"fmt"
	"net/http"

	"github.com/newaccg/todo-api/internal/constants"
)

func (h *handler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	hdr := h.midware.RateLimit(h.Register, "/register", h.bucketSizes.Register)
	mux.HandleFunc("POST /register", Handle(hdr))

	hdr = h.midware.RateLimit(h.Login, "/login", h.bucketSizes.Login)
	mux.HandleFunc("POST /login", Handle(hdr))

	hdr = h.midware.RateLimit(h.Refresh, "/refresh", h.bucketSizes.Refresh)
	mux.HandleFunc("POST /refresh", Handle(hdr))

	hdr = h.midware.RateLimit(h.GetTasks, "/todos", h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc("GET /todos", Handle(hdr))

	hdr = h.midware.RateLimit(h.CreateTask, "/todos", h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc("POST /todos", Handle(hdr))

	hdr = h.midware.RateLimit(h.UpdateTask, "/todos", h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc(fmt.Sprintf("PUT /todos/{%s}", constants.URLIDValue), Handle(hdr))

	hdr = h.midware.RateLimit(h.DeleteTask, "/todos", h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc(fmt.Sprintf("DELETE /todos/{%s}", constants.URLIDValue), Handle(hdr))

	return mux
}
