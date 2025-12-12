package myjwt

import (
	"errors"
	"io/ioutil"
	"log"
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

func checkAndRefreshTokens(oldAuthTokenString, oldRefreshTokenString, oldCsrfSecretString string) (newRefreshTokenString, newAuthTokenString, newCsrfSecretString string, err error) {
	// Check if csrf secret was provided
	// if not return error right away
	if oldCsrfSecretString == "" {
		log.Println("No CSRF Token")
		err := errors.New("Unauthorized")
		return "", "", "", err
	}
	/*
		What is being done here?
		1. The old token string is being parsed. It contains all information once decoded which need to go into models.TokenClaims()
		2. The ParseWithClaims function also takes model struct as an argument, it will create a struct and fill the struct with details from claims
		3. The third function is a key check function which checks the token against the public key. The function returns
		the public key as verify key so the ParseWithClaims can use that key for verification. The algorithm will be in the
		claims of the token itself
	*/
	authToken, err := jwt.ParseWithClaims(oldAuthTokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	/*
		The line:
		    authTokenClaims, ok := authToken.Claims.(*models.TokenClaims)
		attempts a type assertion. It says:
		    "Treat authToken.Claims as a *models.TokenClaims"
		If the JWT was created using our TokenClaims struct, this will succeed.
		If `ok` is false:
		    - It means the claims inside the JWT did NOT match our expected struct.
		    - This can happen if:
		          - The token was forged with wrong structure
		          - The token uses unexpected claim types
		          - The parsing failed silently
		    - In such a case, we cannot trust the token, so we immediately return
		      from the function (effectively treating the request as unauthorized).
	*/
	authTokenClaims, ok := authToken.Claims.(*models.TokenClaims)
	if !ok {
		return
	}

	/*
		CSRF Secret comes from two different places. One the user from frontend provides CSRF on every request to backend
		via headers. Second the JWT tokens also have csrf in them.
		This line of code validates if the Csrf recieved from the user request matches the csrf in the header of the request recieved via JWT.
		Why Comapre Them?
			- The request came from the legitimate browser session
			- The user intentionally triggered this request
			- Not from another website trying to forge the request (CSRF attack)
	*/
	if oldCsrfSecretString != authTokenClaims.Csrf {
		log.Println("CSRF Token does not match jwt!")
		err = errors.New("Unauthorized")
		return
	}
	/*
		If we reach this branch, it means the AuthToken has been successfully parsed AND is still within its valid lifetime. The
		signature is correct, the structure matches our expected claims, and it has not expired.
	*/
	/*
		All token claims also have a "Valid" boolean. We are checking if it is true
		If it is true we will
		- extract the csrf secret from the auth token claims
			- This is a stable CSRF Secret design. Reusing the CSRF Secret throughout one session is best practice
		- generate a new refresh token with extended expiry
		- keep the auth token same as old one

	*/
	if authToken.Valid {
		log.Println("Auth token is valid")
		newCsrfSecretString = authTokenClaims.Csrf
		newRefreshTokenString, err = updateRefreshTokenExp(oldRefreshTokenString)
		if err != nil {
			return "", "", "", err
		}
		newAuthTokenString = oldAuthTokenString
		return

	}

	return
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
