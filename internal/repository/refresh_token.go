package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

func (r *repository) UpdateRefreshToken(ctx context.Context, userID int64, oldToken, newToken *model.Token) error {
	err := r.validateRefreshToken(ctx, userID, oldToken.Token)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	oldHash := r.crypt.StringToSha256(oldToken.Token)
	newHash := r.crypt.StringToSha256(newToken.Token)

	_, err = tx.ExecContext(ctx,
		"UPDATE refresh_tokens SET token_hash = ?, expires_at = ? WHERE user_id = ? AND token_hash = ?",
		newHash,
		newToken.ExpirationTime,
		userID,
		oldHash,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) InsertRefreshToken(ctx context.Context, refreshToken *model.Token, userID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hash := r.crypt.StringToSha256(refreshToken.Token)

	_, err = tx.ExecContext(ctx,
		"INSERT INTO refresh_tokens (user_id, expires_at, token_hash) VALUES (?, ?, ?)",
		userID,
		refreshToken.ExpirationTime,
		hash,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) validateRefreshToken(ctx context.Context, userID int64, token string) error {
	var exp int64
	hash := r.crypt.StringToSha256(token)

	row := r.db.QueryRowContext(ctx,
		"SELECT expires_at FROM refresh_tokens WHERE user_id = ? AND token_hash = ?",
		userID,
		hash,
	)

	err := row.Scan(&exp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrInvalidToken
		}

		return err
	}

	if time.Now().Unix() > exp {
		r.deleteRefreshToken(ctx, userID, token)
		return errs.ErrTokenExpired
	}

	return nil
}

func (r *repository) deleteRefreshToken(ctx context.Context, userID int64, token string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"DELETE FROM refresh_tokens WHERE user_id = ? AND token_hash = ?",
		userID,
		token,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
