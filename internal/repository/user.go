package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

	errs "github.com/newaccg/todo-api/internal/errors"
)

func (r *repository) Register(ctx context.Context, name, email, passwordHash string) (int64, error) {
	// transaction for rollback
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// inserting user
	res, err := tx.ExecContext(ctx,
		"INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
		name,
		email,
		passwordHash,
	)
	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok {
			// error 1062 indicates that we tried to insert a duplicate of a unique value
			if sqlErr.Number == 1062 {
				return 0, errs.ErrEmailExists
			}
		}

		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) GetUserIDAndPasswordHashByEmail(ctx context.Context, email string) (int64, string, error) {
	var id int64
	var hash string

	err := r.db.QueryRowContext(ctx,
		"SELECT id, password FROM users WHERE email = ?",
		email,
	).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", errs.ErrUserNotFound
		}

		return 0, "", err
	}

	return id, hash, nil
}
