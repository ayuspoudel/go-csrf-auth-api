/*
Author: Ayush Poudel

This package defines GenerateRandomStrings functions which will be used to generate a CSRF Secret.
It uses crypto/rand and encoding/base64 packages to generate secure random strings.
The steps here are:
	1. GenerateRandomBytes function generates n random bytes using crypto/rand's Read function.
	2. GenerateRandomStrings function encodes those random bytes into a URL-safe base64 string using encoding/base64's URLEncoding.EncodeToString function.

*/

package randomStrings

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomStrings(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	randomString := base64.URLEncoding.EncodeToString(b)
	return randomString, err
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
