package repository

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/go-sql-driver/mysql"

	"github.com/newaccg/todo-api/internal/config"
)

type crypto interface {
	EncryptPassword(str string) (string, error)
	ArePasswordAndHashEqual(str string, hash string) (bool, error)
	StringToSha256(str string) string
}

type repository struct {
	db    *sql.DB
	crypt crypto

	orders *config.OrdersConfig
	config *config.DBConfig
}

func NewRepository(config *config.DBConfig, cryp crypto, ord *config.OrdersConfig) *repository {
	return &repository{
		config: config,
		crypt:  cryp,
		orders: ord,
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

	// check status of the event scheduler
	var status string
	err = db.QueryRow("SELECT @@global.event_scheduler").Scan(&status)
	if err != nil {
		log.Fatal(err)
	}

	if status != "ON" {
		return errors.New("event_scheduler is not set to ON. If you use Linux, please set event_scheduler=ON in the [mysqld] section in mysql configuration file")
	}

	r.db = db
	return nil
}

func (r *repository) UnloadDB() {
	if err := r.db.Close(); err != nil {
		slog.Error(
			"could not unload DB:",
			"error", err,
		)
	}
}
