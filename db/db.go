package db

import (
	"errors"

	"github.com/ayuspoudel/go-csrf-auth-api/db/models"
)

var users = map[string]models.User{}

func InitializeDB() {

}

func StoreUser(username, password, role string) (uuid string, err error) {

	return
}
func DeleteUser() {

}
func FetchUserByUserName(username string) (models.User, string, error) {
	for k, user := range users {
		if user.Username == username {
			return user, k, nil
		}
	}
	return models.User{}, "", errors.New("user not found that matches the given username")
}

func FetchUserById() {

}

func StoreRefreshToken() {
	return
}

func DeleteRefreshToken() {

}

func CheckRefreshToken() bool {
	return true
}

func logUserIn() {

}

func generateByCryptHash() {

}

func checkPasswordAgainstHash() {

}

func logUserOut() {

}
