package myjwt

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"log"
	"time"

	"github.com/ayuspoudel/go-csrf-auth-api/db"
	"github.com/ayuspoudel/go-csrf-auth-api/db/models"
	"github.com/dgrijalva/jwt-go"
)

const privKeyPath = "keys/app.rsa"
const pubKeyPath = "keys/app.rsa.pub"

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

func InitJWT() error {
	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		return err
	}
	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		return err
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return err
	}
	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
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
		newRefreshTokenString, err = updateRefreshTokenExpiry(oldRefreshTokenString)
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

func createRefreshTokenString(uuid, role, csrfSecret string) (refreshTokenString string, err error) {

	// Get a new timestamp
	refreshTokenExp := time.Now().Add(models.RefreshTokenValidTime).Unix()
	// Get a new ID as JTI and store it in DB as valid "isdoubfoiw382r = valid"
	refreshJti, err := db.StoreRefreshToken()
	if err != nil {
		log.Panic("panic: %+v", err)
		return
	}
	// Pass the jti, expiry time, csrf, role and uuid into our tokenclaims struct
	refreshClaims := models.TokenClaims{
		jwt.StandardClaims{
			Id:        refreshJti,
			Subject:   uuid,
			ExpiresAt: refreshTokenExp,
			IssuedAt:  time.Now().Unix(),
		},
		role,
		csrfSecret,
	}
	// Pass the token claims struct we just build to NewWithClaims and build a token with RS256 Algorithm
	refreshJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), refreshClaims)
	// Get token string and error if any from the refreshJwt and return
	refreshTokenString, err = refreshJwt.SignedString(signKey)
	return
}

func updateRefreshTokenExpiry(oldRefreshTokenString string) (newRefreshToken string, err error) {
	// Get refreshtoken from the old token string with verify key
	refreshToken, err := jwt.ParseWithClaims(oldRefreshTokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	if err != nil {
		return
	}
	//Type cast the old claim into our models.TokenClaims struct
	oldRefreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
	if !ok {
		return
	}
	// Create a new refresh token string with old claims but extended expiry time
	newRefreshToken, err = createRefreshTokenString(oldRefreshTokenClaims.Subject, oldRefreshTokenClaims.Role, oldRefreshTokenClaims.Csrf)
	return
}

func updateAuthTokenString(refreshTokenString string, oldAuthTokenString string) (newAuthTokenString, csrfSecret string, err error) {
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	if err != nil {
		return
	}
	refreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
	if !ok {
		return
	}
	if db.CheckRefreshToken(refreshTokenClaims.StandardClaims.Id) {
		if refreshToken.Valid {
			oldAuthToken, _ := jwt.ParseWithClaims(oldAuthTokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
				return verifyKey, nil
			})
			oldAuthTokenClaims, ok := oldAuthToken.Claims.(*models.TokenClaims)
			if !ok {
				return
			}
			csrfSecret, err = models.GenerateCSRFSecret()
			if err != nil {
				log.Println("Error creating CSRF Secret: %+v", err)
				return
			}
			newAuthTokenString, err = createAuthTokenString(oldAuthTokenClaims.StandardClaims.Subject, oldAuthTokenClaims.Role, csrfSecret)
			return
		} else {
			log.Println("Refresh Token has expired")
			db.DeleteRefreshToken(refreshTokenClaims.StandardClaims.Id)
			err = errors.New("Unauthorized")
			return
		}
	} else {
		log.Println("Refresh Token has been revoked")
		err = errors.New("Unauthorized")
		return
	}
}

func revokeRefreshToken(refreshTokenString string) error {
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})
	if err != nil {
		return errors.New("could not parse refresh token with claims")
	}
	refreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
	if !ok {
		return errors.New("could not type assert refresh token claims")
	}
	db.DeleteRefreshToken(refreshTokenClaims.StandardClaims.Id)
	return nil

}

func updateRefreshTokenCsrf(oldRefreshTokenString string, newCsrfString string) (newRefreshTokenString string, err error) {
	refreshToken, err := jwt.ParseWithClaims(oldRefreshTokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	oldRefreshTokenClaims, ok := refreshToken.Claims.(*models.TokenClaims)
	if !ok {
		return
	}

	refreshClaims := models.TokenClaims{
		jwt.StandardClaims{
			Id:        oldRefreshTokenClaims.StandardClaims.Id, // jti
			Subject:   oldRefreshTokenClaims.StandardClaims.Subject,
			ExpiresAt: oldRefreshTokenClaims.StandardClaims.ExpiresAt,
		},
		oldRefreshTokenClaims.Role,
		newCsrfString,
	}

	refreshJwt := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), refreshClaims)

	newRefreshTokenString, err = refreshJwt.SignedString(signKey)
	return

}

func GrabUUID(authTokenString string) (string, error) {
	authToken, _ := jwt.ParseWithClaims(authTokenString, &models.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return "", errors.New("Error fetching claims")
	})
	authTokenClaims, ok := authToken.Claims.(*models.TokenClaims)
	if !ok {
		return "", errors.New("Error fetching claims")
	}

	return authTokenClaims.StandardClaims.Subject, nil
}
