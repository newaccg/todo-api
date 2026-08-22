package model

type User struct {
	ID       int64
	Name     string
	Email    string
	Password string
}

type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type TokenPair struct {
	RefreshToken Token `json:"refreshToken"`
	AccessToken  Token `json:"accessToken"`
}

type Token struct {
	Token          string
	ExpirationTime int64 // UNIX time
}

type Claims struct {
	ExpirationTime int64 // UNIX time
	UserID         int64
}
