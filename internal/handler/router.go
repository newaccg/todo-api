package handler

import "net/http"

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
	mux.HandleFunc("PUT /todos/{id}", Handle(hdr))

	hdr = h.midware.RateLimit(h.DeleteTask, "/todos", h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc("DELETE /todos/{id}", Handle(hdr))

	return mux
}
