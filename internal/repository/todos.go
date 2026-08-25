package repository

import (
	"context"

	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

func (r *repository) CreateWithUserID(ctx context.Context, task *model.Task, userID int64) (*model.Task, error) {
	// transaction for rollback
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// getting future task ID
	var id int64
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(id), 0) + 1
		from todos
		WHERE user_id = ?
		FOR UPDATE
	`, userID,
	).Scan(&id)

	// inserting task
	_, err = tx.ExecContext(ctx,
		`
		INSERT INTO todos (id, user_id, title, description)
		VALUES (?, ?, ?, ?);
		`,
		id,
		userID,
		task.Title,
		task.Description,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	task.ID = id

	return task, nil
}

func (r *repository) GetAllWithUserID(ctx context.Context, userID int64, filter, order string, page, limit int) ([]model.Task, error) {
	// filtering
	filter = "%" + filter + "%"

	// common arguments for query
	args := []any{
		userID,
		filter,
		filter,
	}

	query := "SELECT id, title, description FROM todos WHERE user_id = ? AND (title LIKE ? OR description LIKE ?)"

	// selecting order
	if order != "" {
		var ord string

		switch order {

		case r.orders.Description:
			ord = "description"

		case r.orders.ID:
			ord = "id"

		case r.orders.Title:
			ord = "title"

		default:
			return nil, errs.ErrInvalidOrder
		}

		query += " ORDER BY " + ord
	}

	// pagination
	if page != 0 && limit != 0 {
		query += " LIMIT ?, ? "

		args = append(args, page*limit-limit)
		args = append(args, limit)
	}

	todos := make([]model.Task, 0)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task
		rows.Scan(&task.ID, &task.Title, &task.Description)

		todos = append(todos, task)
	}

	return todos, nil
}

func (r *repository) DeleteByIDWithUserID(ctx context.Context, taskID, userID int64) error {
	// transaction for rollback
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := r.db.ExecContext(ctx, "DELETE FROM todos WHERE id = ? AND user_id = ?", taskID, userID)
	if err != nil {
		return err
	}

	deletedCount, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if deletedCount == 0 {
		return errs.ErrTaskNotFound
	}

	return tx.Commit()
}

func (r *repository) UpdateByIDWithUserID(ctx context.Context, taskID, userID int64, task *model.Task) (*model.Task, error) {
	// transaction for rollback
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// updating task
	res, err := tx.ExecContext(ctx,
		"UPDATE todos SET title = ?, description = ? WHERE id = ? AND user_id = ?",
		task.Title,
		task.Description,
		taskID,
		userID,
	)

	if err != nil {
		return nil, err
	}

	updatedCount, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if updatedCount == 0 { // if no rows affected...
		// task with this ID doesn't exist
		return nil, errs.ErrTaskNotFound
	}

	task.ID = taskID

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return task, nil
}
