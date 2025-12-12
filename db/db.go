package db

import (
	"errors"
	"log"

	"github.com/ayuspoudel/go-csrf-auth-api/db/models"
	"github.com/ayuspoudel/go-csrf-auth-api/randomStrings"
	"golang.org/x/crypto/bcrypt"
)

var users  = map[string]models.User
var refreshTokens map[string]string

func InitializeDB() {
	refreshTokens := make(map[string]string)
}

func StoreUser(username, password, role string) (uuid string, err error) {
	uuid, err = randomStrings.GenerateRandomStrings(32)
	if err != nil {
		return "", err
	}
	u := models.User{}
	for u != users[uuid] { //users[uuid] will return a blank user model if not found
		uuid, err = randomStrings.GenerateRandomStrings(32)
		if err != nil {
			return "", err
		}
	}
	passwordHash, hasherr := generateBcryptHash(password)
	if hasherr != nil {
		err = hasherr
		return "", err
	}
	users[uuid] = models.User{username, passwordHash, role}

	return uuid, err
}

func DeleteUser(uuid string) {
	delete(users, uuid)
}

func FetchUserByUserName(username string) (models.User, string, error) {
	for k, user := range users {
		if user.Username == username {
			return user, k, nil
		}
	}
	return models.User{}, "", errors.New("user not found that matches the given username")
}

func FetchUserById(uuid string) (models.User, error) {
	u := users[uuid] // This will return a blank user if nothing is found
	blankUser := models.User{}
	if blankUser != u {
		return u, nil
	}
	return u, errors.New("User not found that matches given username")
}

func StoreRefreshToken() (jti string, err error) {
	jti, err = randomStrings.GenerateRandomStrings(32)
	if err != nil {
		return "", err
	}
	for refreshTokens[jti] != "" { //refreshTokens[jti] will return empty string if is unique
		jti, err = randomStrings.GenerateRandomStrings(32)
		return jti, err
	}
	refreshTokens[jti] = "valid"
	return jti, err

}

func DeleteRefreshToken(jti string) {
	delete(refreshTokens, jti)
}

func CheckRefreshToken(jti string) bool {
	return refreshTokens[jti] != 
}

func logUserIn(username string, password string) (models.User, string, error) {
	user, uuid, err := FetchUserByUserName(username)
	log.Println(user, uuid, err)
	if err != nil {
		return models.User{}, "", err
	}
	err = checkPasswordAgainstHash(user.PasswordHash, password)
	return user, uuid, err

}

func generateBcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash[:]), err
}

func checkPasswordAgainstHash(hash string, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func logUserOut() {

}
