package service

import (
	"context"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

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

	refresh, err := s.encryptRefreshTokenFromPair(newTokens)
	if err != nil {
		return nil, err
	}

	err = s.repo.UpdateRefreshToken(ctx, oldToken, refresh)
	if err != nil {
		return nil, err
	}

	return newTokens, nil
}

func (s *Service) generateAndInsertTokens(ctx context.Context, userID int64) (*model.TokenPair, error) {
	tokens, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	refresh, err := s.encryptRefreshTokenFromPair(tokens)
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

func (s *Service) encryptRefreshTokenFromPair(pair *model.TokenPair) (*model.Token, error) {
	refresh := pair.RefreshToken

	token, err := s.crypt.Encrypt(refresh.Token)
	if err != nil {
		return nil, err
	}

	refresh.Token = token

	return &refresh, nil
}
