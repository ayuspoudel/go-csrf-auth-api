package models

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type User struct {
	Username     string
	PasswordHash string
	Role         string
}

type TokenClaims struct {
	jwt.StandardClaims
	Role string `json:"role"`
	Csrf string `json:"csrf"`
}

const refreshTokenValidTime = time.Hour * 72
const AuthTokenValidTime = time.Minute * 15

func GenerateCSRFSecret() (csrf string, err error) {
	csrf = randomstrings.GenerateRandomString[15]
}
