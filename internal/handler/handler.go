package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

type Service interface {
	RegisterUser(ctx context.Context, name, email, password string) (string, error)
	LoginUser(ctx context.Context, email, password string) (string, error)
	GetAllTasksWithUserID(ctx context.Context, userID int64) ([]model.Task, error)
	GetTasksByFilterWithUserID(ctx context.Context, userID int64, filter string) ([]model.Task, error)
	PaginateTasks(tasks []model.Task, page, limit int) ([]model.Task, error)
	CreateTaskWithUserID(ctx context.Context, title, description string, userId int64) (*model.Task, error)
	UpdateTaskByIDWithUserID(ctx context.Context, taskID, userId int64, title, description string) (*model.Task, error)
	DeleteTaskByIDWithUserID(ctx context.Context, taskID, userId int64) error
}

type middleware interface {
	Auth(CustomHandler) CustomHandler
}

type handler struct {
	service Service
	midware middleware
}

type taskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewHandler(svc Service, mware middleware) *handler {
	return &handler{
		service: svc,
		midware: mware,
	}
}

func (h *handler) RegisterRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", Handle(h.Register))
	mux.HandleFunc("POST /login", Handle(h.Login))

	mux.HandleFunc("GET /todos", Handle(h.midware.Auth(h.GetTasks)))
	mux.HandleFunc("POST /todos", Handle(h.midware.Auth(h.CreateTask)))

	mux.HandleFunc("PUT /todos/{id}", Handle(h.midware.Auth(h.UpdateTask)))
	mux.HandleFunc("DELETE /todos/{id}", Handle(h.midware.Auth(h.DeleteTask)))

	return mux
}

func (h *handler) Register(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errs.ErrInvalidJSON
	}

	if err := validateInputStrings(input.Name, input.Email, input.Password); err != nil {
		return err
	}

	token, err := h.service.RegisterUser(r.Context(), input.Name, input.Email, input.Password)
	if err != nil {
		return err
	}

	if err := writeToken(w, token); err != nil {
		return err
	}

	return nil
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errs.ErrInvalidJSON
	}

	if err := validateInputStrings(input.Email, input.Password); err != nil {
		return err
	}

	token, err := h.service.LoginUser(r.Context(), input.Email, input.Password)
	if err != nil {
		return err
	}

	if err := writeToken(w, token); err != nil {
		return err
	}

	return nil
}

func (h *handler) GetTasks(w http.ResponseWriter, r *http.Request) error {
	id, err := getUserIDFromContext(r.Context())
	if err != nil {
		return err
	}

	var tasks []model.Task

	query := r.URL.Query()
	filter := query.Get("term")
	ctx := r.Context()

	if filter == "" { // if filter is specified...
		tasks, err = h.service.GetAllTasksWithUserID(ctx, id)
	} else {
		tasks, err = h.service.GetTasksByFilterWithUserID(ctx, id, filter)
	}
	if err != nil {
		return err
	}

	bad := false

	page, err := getIntFromURLQuery(query, "page")
	if err != nil {
		bad = true
	}

	limit, err := getIntFromURLQuery(query, "limit")
	if err != nil {
		bad = true
	}

	if !bad { // if page or limit values are not invalid...
		// return paginated tasks
		tasks, err = h.service.PaginateTasks(tasks, page, limit)
		if err != nil {
			return err
		}

		output := struct {
			Data  []model.Task `json:"data"`
			Page  int          `json:"page"`
			Limit int          `json:"limit"`
			Total int          `json:"total"`
		}{
			Data:  tasks,
			Page:  page,
			Limit: limit,
			Total: len(tasks),
		}

		writeJSON(w, output)
	} else {
		// return common tasks
		writeJSON(w, tasks)
	}

	return nil
}

func (h *handler) CreateTask(w http.ResponseWriter, r *http.Request) error {
	var input taskRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errs.ErrInvalidJSON
	}

	ctx := r.Context()

	id, err := getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	task, err := h.service.CreateTaskWithUserID(ctx, input.Title, input.Description, id)
	if err != nil {
		return err
	}

	writeJSONWithCode(w, task, http.StatusCreated)
	return nil
}

func (h *handler) UpdateTask(w http.ResponseWriter, r *http.Request) error {
	var input taskRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return errs.ErrInvalidJSON
	}

	taskID, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	ctx := r.Context()

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	task, err := h.service.UpdateTaskByIDWithUserID(ctx, taskID, userID, input.Title, input.Description)
	if err != nil {
		return err
	}

	writeJSON(w, task)
	return nil
}

func (h *handler) DeleteTask(w http.ResponseWriter, r *http.Request) error {
	taskID, err := getIDFromRequest(r)
	if err != nil {
		return err
	}

	ctx := r.Context()

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	err = h.service.DeleteTaskByIDWithUserID(ctx, taskID, userID)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

// structural validation
func validateInputStrings(input ...string) error {
	if slices.Contains(input, "") {
		return errs.ErrEmptyField
	}

	return nil
}

func getUserIDFromContext(ctx context.Context) (int64, error) {
	// we cannot convert directly to int64 because jwt.MapClaims converts all numeric values to float64
	// so we have to do this intermediate step
	floatID, ok := ctx.Value("userID").(float64)
	if !ok {
		return 0, errors.New("could not convert ID from context to float64")
	}

	// this is guaranteed to be converted from float64 to int64
	// so we don't have to check for success
	id := int64(floatID)

	return id, nil
}

func getIDFromRequest(r *http.Request) (int64, error) {
	str := r.PathValue("id")
	id, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, errs.ErrInvalidID
	}

	// structural validation
	if id <= 0 {
		return 0, errs.ErrInvalidID
	}

	return id, nil
}

func getIntFromURLQuery(query url.Values, value string) (int, error) {
	str := query.Get(value)
	res, err := strconv.Atoi(str)
	if err != nil || res <= 0 {
		return 0, errs.ErrInvalidURLValue
	}

	return res, nil
}

func writeJSON(w http.ResponseWriter, msg any) {
	writeJSONWithCode(w, msg, http.StatusOK)
}

func writeToken(w http.ResponseWriter, token string) error {
	var output struct {
		Token string `json:"token"`
	}
	output.Token = token

	if err := json.NewEncoder(w).Encode(&output); err != nil {
		return err
	}

	return nil
}
