package model

type User struct {
	ID       int64
	Name     string
	Email    string
	Password string
}

type Task struct {
	ID          int64 `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
