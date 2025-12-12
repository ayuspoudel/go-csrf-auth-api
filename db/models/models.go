/*
Author: Ayush poudel

This is models package which defines struct for User and TokenClaims to be used for JWT Auth.
User model is simple contains UserName and Password and Role, which should be enhanced later using struct tags for validation of password and role types used in the api.
TokenClaims struct embeds jwt.StandardClaims and adds Role and Csrf fields to store user role and CSRF token respectively.

It also defines constants for refresh token validity time and auth token validity time.
The GenerateCSRFSecret function generates a secure random CSRF secret using the randomStrings package.

*/

package models

import (
	"time"

	randomStrings "github.com/ayuspoudel/go-csrf-auth-api/randomStrings"
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

const RefreshTokenValidTime = time.Hour * 72
const AuthTokenValidTime = time.Minute * 15

func GenerateCSRFSecret() (csrf string, err error) {
	csrf, err = randomStrings.GenerateRandomStrings(32)
	return csrf, err
}
