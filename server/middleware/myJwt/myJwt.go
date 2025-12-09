package myjwt

import (
	"io/ioutil"
	"time"

	"github.com/ayuspoudel/go-csrf-auth-api/db/models"
	"github.com/dgrijalva/jwt-go"
)

const privKeyPath = "keys/app.rsa"
const pubKeyPath = "keys/app.rsa.pub"

func InitJWT() error {
	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return err
	}
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return err
	}
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return err
	}
	return nil
}

func CreateNewTokens(uuid, role string) (authToken, refreshToken, csrfSecret string, err error) {
	// generate CSRF secret
	//generate refresh token
	// generate auth token

	csrfSecret, err = models.GenerateCSRFSecret()
	if err != nil {
		return
	}
	refreshToken, err = createAuthTokenString(uuid, role, csrfSecret)
	if err != nil {
		return
	}
	authToken, err = createAuthTokenString(uuid, role, csrfSecret)
	if err != nil {
		return
	}
	return

}

func checkAndRefreshTokens(uuid, role, csrfSecret string) (refreshToken string, err error) {

}

func createAuthTokenString(uuid, role, csrfSecret string) (authTokenString string, err error) {
	authTokenExp := time.Now().Add(models.AuthTokenValidTime).Unix()
	authClaims := models.TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: authTokenExp,
			IssuedAt:  time.Now().Unix(),
			Subject:   uuid,
		},
		Role: role,
		Csrf: csrfSecret,
	}
	authJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), authClaims)
	authTokenString, err = authJwt.SignedString(signKey)
	return
}

func createRefreshTokenString() {

}

func updateRefreshTokenExpiry() {

}

func updateAuthTokenString() {

}

func revokeRefreshToken() error {

}

func GrabUUID() {

}
