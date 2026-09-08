package service

import (
	"context"

	"github.com/newaccg/todo-api/internal/config"
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
