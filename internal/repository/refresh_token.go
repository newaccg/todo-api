package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

func (r *repository) UpdateRefreshToken(ctx context.Context, oldToken, newToken *model.Token) error {
	err := r.validateRefreshToken(ctx, oldToken)
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"DELETE FROM refresh_tokens WHERE id = ?",
		oldToken.Claims.TokenID,
	)
	if err != nil {
		return err
	}

	err = r.insertRefreshToken(ctx, newToken, tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) InsertRefreshToken(ctx context.Context, refreshToken *model.Token) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = r.insertRefreshToken(ctx, refreshToken, tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) insertRefreshToken(ctx context.Context, refreshToken *model.Token, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx,
		"INSERT INTO refresh_tokens (user_id, id, expires_at, token_hash) VALUES (?, ?, ?, ?)",
		refreshToken.Claims.UserID,
		refreshToken.Claims.TokenID,
		refreshToken.Claims.ExpirationTime,
		refreshToken.Token,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) validateRefreshToken(ctx context.Context, token *model.Token) error {
	var exp int64

	row := r.db.QueryRowContext(ctx,
		"SELECT expires_at FROM refresh_tokens WHERE id = ?",
		token.Claims.TokenID,
	)

	err := row.Scan(&exp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrInvalidToken
		}

		return err
	}

	if time.Now().Unix() > exp {
		r.deleteRefreshToken(ctx, token)
		return errs.ErrTokenExpired
	}

	return nil
}

func (r *repository) deleteRefreshToken(ctx context.Context, token *model.Token) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"DELETE FROM refresh_tokens WHERE id = ?",
		token.Claims.UserID,
		token.Token,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
