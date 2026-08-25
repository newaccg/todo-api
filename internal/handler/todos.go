package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

func (h *handler) GetTasks(w http.ResponseWriter, r *http.Request) error {
	id, err := h.getUserIDFromContext(r.Context())
	if err != nil {
		return err
	}

	var tasks []model.Task

	query := r.URL.Query()
	filter := query.Get(h.urlValueNames.Filter)
	order := query.Get(h.urlValueNames.Order)
	ctx := r.Context()

	page, err := getIntFromURLQuery(query, h.urlValueNames.Page)
	if err != nil {
		return err
	}

	limit, err := getIntFromURLQuery(query, h.urlValueNames.Limit)
	if err != nil {
		return err
	}

	tasks, err = h.service.GetAllTasksWithUserID(ctx, id, filter, order, page, limit)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidOrder) {
			return errs.ErrInvalidURLValue
		}

		return err
	}

	if page+limit != 0 && page*limit == 0 { // if either page or limit are not specified...
		return errs.ErrInvalidURLValue
	}

	if page*limit != 0 { // if page and limit values are specified...
		// return paginated tasks
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

	id, err := h.getUserIDFromContext(ctx)
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

	userID, err := h.getUserIDFromContext(ctx)
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

	userID, err := h.getUserIDFromContext(ctx)
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

func getIntFromURLQuery(query url.Values, value string) (int, error) {
	str := query.Get(value)
	if str == "" {
		return 0, nil
	}

	res, err := strconv.Atoi(str)
	if err != nil || res <= 0 {
		return 0, errs.ErrInvalidURLValue
	}

	return res, nil
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

func (h *handler) getUserIDFromContext(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(h.jwtUserIDValueName).(int64)
	if !ok {
		return 0, errors.New("invalid JWT user ID value: could not convert to int64")
	}

	return id, nil
}
