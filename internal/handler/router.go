package handler

import "net/http"

func (h *handler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	hdr := h.midware.RateLimit(h.Register, h.bucketSizes.Register)
	mux.HandleFunc("POST /register", Handle(hdr))

	hdr = h.midware.RateLimit(h.Login, h.bucketSizes.Login)
	mux.HandleFunc("POST /login", Handle(hdr))

	hdr = h.midware.RateLimit(h.Refresh, h.bucketSizes.Refresh)
	mux.HandleFunc("POST /refresh", Handle(hdr))

	hdr = h.midware.RateLimit(h.GetTasks, h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc("GET /todos", Handle(hdr))

	hdr = h.midware.RateLimit(h.CreateTask, h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc("POST /todos", Handle(hdr))

	hdr = h.midware.RateLimit(h.UpdateTask, h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc("PUT /todos/{id}", Handle(hdr))

	hdr = h.midware.RateLimit(h.DeleteTask, h.bucketSizes.Todos)
	hdr = h.midware.Auth(hdr)
	mux.HandleFunc("DELETE /todos/{id}", Handle(hdr))

	return mux
}
