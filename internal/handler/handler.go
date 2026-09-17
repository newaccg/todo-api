package handler

import (
	"context"
	"net/http"

	"github.com/newaccg/todo-api/internal/config"
	"github.com/newaccg/todo-api/internal/model"
)

type Service interface {
	RegisterUser(ctx context.Context, name, email, password string) (*model.TokenPair, error)
	LoginUser(ctx context.Context, email, password string) (*model.TokenPair, error)
	GetAllTasksWithUserID(ctx context.Context, userID int64, filter, order string, page, limit int) ([]model.Task, error)
	CreateTaskWithUserID(ctx context.Context, title, description string, userId int64) (*model.Task, error)
	UpdateTaskByIDWithUserID(ctx context.Context, taskID, userId int64, title, description string) (*model.Task, error)
	UpdateRefreshToken(ctx context.Context, oldToken string) (*model.TokenPair, error)
	DeleteTaskByIDWithUserID(ctx context.Context, taskID, userId int64) error
}

type middleware interface {
	Auth(CustomHandler) CustomHandler
	RateLimit(next CustomHandler, path string, bucketSize int) CustomHandler
}

type handler struct {
	urlValueNames      *config.ValueNamesURL
	bucketSizes        *config.BucketSizesConfig

	service Service
	midware middleware
}

type taskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewHandler(svc Service, mware middleware, urlVals *config.ValueNamesURL, sizes *config.BucketSizesConfig) *handler {
	return &handler{
		service:            svc,
		midware:            mware,
		urlValueNames:      urlVals,
		bucketSizes:        sizes,
	}
}

func writeJSON(w http.ResponseWriter, msg any) {
	writeJSONWithCode(w, msg, http.StatusOK)
}
