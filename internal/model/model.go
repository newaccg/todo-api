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
	Token  string
	Claims Claims
}

type Claims struct {
	TokenID        string
	ExpirationTime int64 // UNIX time
	UserID         int64
}
