package service

import (
	"context"
	"fmt"

	"github.com/newaccg/todo-api/internal/config"
	"github.com/newaccg/todo-api/internal/model"
)

type Repository interface {
	Register(ctx context.Context, name, email, password string) (int64, error)
	Login(ctx context.Context, email, password string) (int64, error)
	GetAllWithUserID(ctx context.Context, userID int64) ([]model.Task, error)
	CreateWithUserID(ctx context.Context, task *model.Task, id int64) (*model.Task, error)
	UpdateByIDWithUserID(ctx context.Context, taskID, userID int64, task *model.Task) (*model.Task, error)
	DeleteByIDWithUserID(ctx context.Context, taskID, userID int64) error
}

type JWT interface {
	GenerateJWT(id int64) (string, error)
}

type Service struct {
	cfg *config.JWT
	repo Repository
	jwt JWT
}

func NewService(repo Repository, jwt JWT, conf *config.JWT) *Service {
	return &Service{
		cfg: conf,
		repo: repo,
		jwt: jwt,
	}
}

func (s *Service) RegisterUser(ctx context.Context, name, email, password string) (string, error) {
	// TODO: hash password
	id, err := s.repo.Register(ctx, name, email, password)
	if err != nil {
		return "", err
	}

	return s.jwt.GenerateJWT(id)
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (string, error) {
	id, err := s.repo.Login(ctx, email, password)
	if err != nil {
		return "", err
	}

	return s.jwt.GenerateJWT(id)
}

func (s *Service) GetAllTasksWithUserID(ctx context.Context, userID int64) ([]model.Task, error) {
	tasks, err := s.repo.GetAllWithUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get all tasks: %w", err)
	}

	return tasks, nil
}

func (s *Service) CreateTaskWithUserID(ctx context.Context, title, description string, id int64) (*model.Task, error) {
	task := model.Task{
		Title:    title,
		Description:     description,
	}

	newTask, err := s.repo.CreateWithUserID(ctx, &task, id)
	if err != nil {
		return nil, fmt.Errorf("could not create task: %w", err)
	}

	return newTask, nil
}

func (s *Service) UpdateTaskByIDWithUserID(ctx context.Context, taskID, userID int64, title, description string) (*model.Task, error) {
	task := model.Task{
		Title:    title,
		Description:     description,
	}

	newTask, err := s.repo.UpdateByIDWithUserID(ctx, taskID, userID, &task)
	if err != nil {
		return nil, fmt.Errorf("could not update task by task ID %d and user ID %d: %w", taskID, userID, err)
	}

	return newTask, nil
}

func (s *Service) DeleteTaskByIDWithUserID(ctx context.Context, taskID, userID int64) error {
	if err := s.repo.DeleteByIDWithUserID(ctx, taskID, userID); err != nil {
		return fmt.Errorf("could not delete task by task ID %d and user ID %d: %w", taskID, userID, err)
	}

	return nil
}


