package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/go-sql-driver/mysql"

	"github.com/newaccg/todo-api/internal/config"
	errs "github.com/newaccg/todo-api/internal/errors"
	"github.com/newaccg/todo-api/internal/model"
)

type crypto interface {
	EncryptPassword(password string) ([]byte, error)
	ArePasswordAndHashEqual(password string, hash string) (bool, error)
}

type repository struct {
	db     *sql.DB
	config *config.DBConfig
	crypt crypto
}

func NewRepository(config *config.DBConfig, cryp crypto) *repository {
	return &repository{
		config: config,
		crypt: cryp,
	}
}

func (r *repository) LoadDB() error {
	cfg := mysql.NewConfig()

	// setting config parameters
	cfg.User = r.config.User
	cfg.Passwd = r.config.Password
	cfg.Addr = r.config.Address
	cfg.DBName = r.config.Name

	// setting necessary parameters
	cfg.MultiStatements = true
	cfg.ClientFoundRows = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}

	// find .sql scripts
	path := filepath.Join(r.config.PathToSQLScripts, "*.sql")
	matches, err := filepath.Glob(path)
	if err != nil {
		return err
	}

	// activate .sql scripts
	for _, filename := range matches {
		script, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		_, err = db.Exec(string(script))
		if err != nil {
			return err
		}
	}

	// check DB
	err = db.Ping()
	if err != nil {
		return err
	}

	r.db = db
	return nil
}

func (r *repository) UnloadDB() {
	if err := r.db.Close(); err != nil {
		log.Println("ERROR: could not unload DB:", err)
	}
}

func (r *repository) Register(ctx context.Context, name, email, password string) (int64, error) {
	// transaction for rollback
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	encryptedPassword, err := r.crypt.EncryptPassword(password)
	if err != nil {
		return 0, err
	}

	// inserting user
	res, err := tx.ExecContext(ctx,
		"INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
		name,
		email,
		encryptedPassword,
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

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *repository) Login(ctx context.Context, email, password string) (int64, error) {
	var id int64
	var hash string

	err := r.db.QueryRowContext(ctx,
		"SELECT id, password FROM users WHERE email = ?",
		email,
	).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.ErrUserNotFound
		}

		return 0, err
	}

	matching, err := r.crypt.ArePasswordAndHashEqual(password, hash)
	if err != nil {
		return 0, err
	}

	if !matching {
		return 0, errs.ErrWrondPassword
	}

	return id, nil
}

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
	log.Println(id)

	return task, nil
}

func (r *repository) GetAllWithUserID(ctx context.Context, userID int64) ([]model.Task, error) {
	todos := make([]model.Task, 0)

	rows, err := r.db.QueryContext(ctx, "SELECT id, title, description name FROM todos WHERE user_id = ?", userID)
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
	log.Println(updatedCount)
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
