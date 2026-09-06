package service

import (
	"context"
	"fmt"

	"github.com/newaccg/todo-api/internal/config"
	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

type Repository interface {
	Register(ctx context.Context, name, email, passwordHash string) (int64, error)
	GetUserIDAndPasswordHashByEmail(ctx context.Context, email string) (int64, string, error)
	GetAllWithUserID(ctx context.Context, userID int64, filter, order string, page, limit int) ([]model.Task, error)
	CreateWithUserID(ctx context.Context, task *model.Task, id int64) (*model.Task, error)
	UpdateByIDWithUserID(ctx context.Context, taskID, userID int64, task *model.Task) (*model.Task, error)
	UpdateRefreshToken(ctx context.Context, oldToken, newToken *model.Token) error
	DeleteByIDWithUserID(ctx context.Context, taskID, userID int64) error
	InsertRefreshToken(ctx context.Context, refreshToken *model.Token) error
}

type JWT interface {
	GenerateAccessToken(userID int64) (*model.Token, error)
	GenerateRefreshToken(userID int64) (*model.Token, error)
	ValidateAndGetClaimsFromJWT(token string) (*model.Claims, error)
}

type crypto interface {
	Encrypt(str string) (string, error)
	AreStringAndHashEqual(str string, hash string) (bool, error)
}

type Service struct {
	cfg *config.JWT

	repo  Repository
	jwt   JWT
	crypt crypto
}

func NewService(repo Repository, jwt JWT, conf *config.JWT, cr crypto) *Service {
	return &Service{
		cfg:   conf,
		repo:  repo,
		jwt:   jwt,
		crypt: cr,
	}
}

func (s *Service) RegisterUser(ctx context.Context, name, email, password string) (*model.TokenPair, error) {
	hash, err := s.crypt.Encrypt(password)
	if err != nil {
		return nil, err
	}

	id, err := s.repo.Register(ctx, name, email, hash)
	if err != nil {
		return nil, err
	}

	return s.generateAndInsertTokens(ctx, id)
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (*model.TokenPair, error) {
	id, hash, err := s.repo.GetUserIDAndPasswordHashByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	match, err := s.crypt.AreStringAndHashEqual(password, hash)
	if err != nil {
		return nil, err
	}

	if !match {
		return nil, errs.ErrWrongPassword
	}

	return s.generateAndInsertTokens(ctx, id)
}

func (s *Service) GetAllTasksWithUserID(ctx context.Context, userID int64, filter, order string, page, limit int) ([]model.Task, error) {
	tasks, err := s.repo.GetAllWithUserID(ctx, userID, filter, order, page, limit)
	if err != nil {
		return nil, fmt.Errorf("could not get all tasks: %w", err)
	}

	return tasks, nil
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

func (s *Service) UpdateRefreshToken(ctx context.Context, oldTokenStr string) (*model.TokenPair, error) {
	claims, err := s.jwt.ValidateAndGetClaimsFromJWT(oldTokenStr)
	if err != nil {
		return nil, err
	}
	userID := claims.UserID

	newTokens, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	oldToken := &model.Token{
		Token:  oldTokenStr,
		Claims: *claims,
	}

	refresh, err := s.ecnryprRefreshTokenFromPair(newTokens)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateRefreshToken(ctx, oldToken, refresh)
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

	refresh, err := s.ecnryprRefreshTokenFromPair(tokens)
	if err != nil {
		return nil, err
	}

	err = s.repo.InsertRefreshToken(ctx, refresh)
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

func (s *Service) ecnryprRefreshTokenFromPair(pair *model.TokenPair) (*model.Token, error) {
	refresh := pair.RefreshToken

	token, err := s.crypt.Encrypt(refresh.Token)
	if err != nil {
		return nil, err
	}

	refresh.Token = token

	return &refresh, nil
}
