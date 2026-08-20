package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Duration struct {
	time.Duration
}

type Config struct {
	ServerAddress string `json:"serverAddress"`

	DB         DBConfig         `json:"database"`
	Jwt        JWT              `json:"JWT"`
	ValueNames ValueNamesConfig `json:"valueNames"`
}

type ValueNamesConfig struct {
	Url ValueNamesURL `json:"URL"`

	JwtUserID string `json:"JWTUserID"`
}

type ValueNamesURL struct {
	Page   string `json:"page"`
	Limit  string `json:"limit"`
	Filter string `json:"filter"`
}

type JWT struct {
	Secret string

	ExpirationTime Duration `json:"expirationTime"`

	HeaderName string `json:"headerName"`
}

type DBConfig struct {
	Password string
	User     string

	Address          string `json:"address"`
	Name             string `json:"name"`
	PathToSQLScripts string `json:"pathToSQLSrcipts"`
}

// custom rules for unmarshaling JSON (string to time)
func (d *Duration) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

func LoadConfig(configPath string) (*Config, error) {
	var config *Config

	err := godotenv.Load()
	if err != nil {
		return nil,
			fmt.Errorf("could not load environment variables: %w. Please make sure that the .env file is located in the project root and mathes the .env.example file",
				err,
			)
	}

	js, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config.json: %w", err)
	}

	err = json.Unmarshal(js, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config.json: %w", err)
	}

	config.DB.User, err = mustGetenv("DB_USER")
	if err != nil {
		return nil, err
	}

	config.DB.Password = os.Getenv("DB_PASSWORD")

	config.Jwt.Secret, err = mustGetenv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	return config, nil
}

func mustGetenv(variable string) (string, error) {
	res := os.Getenv(variable)
	if res == "" {
		return "", errors.New(variable + " is not specified")
	}

	return res, nil
}
