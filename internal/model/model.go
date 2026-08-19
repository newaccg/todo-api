package model

type User struct {
	ID       int64
	Name     string
	Email    string
	Password string
}

type Task struct {
	ID          int64
	Title       string
	Description string
}
