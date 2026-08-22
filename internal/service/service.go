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
	GetByFilterWithUserID(ctx context.Context, userID int64, filter string) ([]model.Task, error)
	CreateWithUserID(ctx context.Context, task *model.Task, id int64) (*model.Task, error)
	UpdateByIDWithUserID(ctx context.Context, taskID, userID int64, task *model.Task) (*model.Task, error)
	UpdateRefreshToken(ctx context.Context, userID int64, oldToken, newToken string) error
	DeleteByIDWithUserID(ctx context.Context, taskID, userID int64) error
	InsertRefreshToken(ctx context.Context, refreshToken string, userID, expiresAt int64) error
}

type JWT interface {
	GenerateAccessToken(userID int64) (*model.Token, error)
	GenerateRefreshToken(userID int64) (*model.Token, error)
	ValidateAndGetClaimsFromJWT(token string) (*model.Claims, error)
}

type Service struct {
	cfg  *config.JWT
	repo Repository
	jwt  JWT
}

func NewService(repo Repository, jwt JWT, conf *config.JWT) *Service {
	return &Service{
		cfg:  conf,
		repo: repo,
		jwt:  jwt,
	}
}

func (s *Service) RegisterUser(ctx context.Context, name, email, password string) (*model.TokenPair, error) {
	id, err := s.repo.Register(ctx, name, email, password)
	if err != nil {
		return nil, err
	}

	return s.generateAndInsertTokens(ctx, id)
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (*model.TokenPair, error) {
	id, err := s.repo.Login(ctx, email, password)
	if err != nil {
		return nil, err
	}

	return s.generateAndInsertTokens(ctx, id)
}

func (s *Service) GetAllTasksWithUserID(ctx context.Context, userID int64) ([]model.Task, error) {
	tasks, err := s.repo.GetAllWithUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get all tasks: %w", err)
	}

	return tasks, nil
}

func (s *Service) GetTasksByFilterWithUserID(ctx context.Context, userID int64, filter string) ([]model.Task, error) {
	tasks, err := s.repo.GetByFilterWithUserID(ctx, userID, filter)
	if err != nil {
		return nil, fmt.Errorf("could not get tasks by filter %s: %w", filter, err)
	}

	return tasks, nil
}

func (s *Service) PaginateTasks(tasks []model.Task, page, limit int) ([]model.Task, error) {
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex > len(tasks) {
		return make([]model.Task, 0), nil
	}

	if endIndex > len(tasks) {
		endIndex = len(tasks) // removed the " - 1" at the end
	}

	return tasks[startIndex:endIndex], nil
}

func (s *Service) CreateTaskWithUserID(ctx context.Context, title, description string, id int64) (*model.Task, error) {
	task := model.Task{
		Title:       title,
		Description: description,
	}

	newTask, err := s.repo.CreateWithUserID(ctx, &task, id)
	if err != nil {
		return nil, fmt.Errorf("could not create task: %w", err)
	}

	return newTask, nil
}

func (s *Service) UpdateTaskByIDWithUserID(ctx context.Context, taskID, userID int64, title, description string) (*model.Task, error) {
	task := model.Task{
		Title:       title,
		Description: description,
	}

	newTask, err := s.repo.UpdateByIDWithUserID(ctx, taskID, userID, &task)
	if err != nil {
		return nil, fmt.Errorf("could not update task by task ID %d and user ID %d: %w", taskID, userID, err)
	}

	return newTask, nil
}

func (s *Service) UpdateRefreshToken(ctx context.Context, oldToken string) (*model.TokenPair, error) {
	claims, err := s.jwt.ValidateAndGetClaimsFromJWT(oldToken)
	if err != nil {
		return nil, err
	}
	userID := claims.UserID

	newTokens, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateRefreshToken(ctx, userID, oldToken, newTokens.RefreshToken.Token)
	if err != nil {
		return nil, err
	}

	return newTokens, nil
}

func (s *Service) DeleteTaskByIDWithUserID(ctx context.Context, taskID, userID int64) error {
	if err := s.repo.DeleteByIDWithUserID(ctx, taskID, userID); err != nil {
		return fmt.Errorf("could not delete task by task ID %d and user ID %d: %w", taskID, userID, err)
	}

	return nil
}

func (s *Service) generateAndInsertTokens(ctx context.Context, userID int64) (*model.TokenPair, error) {
	tokens, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	err = s.repo.InsertRefreshToken(ctx, tokens.RefreshToken.Token, userID, tokens.RefreshToken.ExpirationTime)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *Service) generateTokenPair(userID int64) (*model.TokenPair, error) {
	access, err := s.jwt.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refresh, err := s.jwt.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	pair := &model.TokenPair{
		AccessToken:  *access,
		RefreshToken: *refresh,
	}

	return pair, nil
}
