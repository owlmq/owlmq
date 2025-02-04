package utils

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"math/big"
)

func GenerateSHA1(str string) string {
	sum := sha1.Sum([]byte(str))
	//this cuts down the hex string to a chosen size
	return fmt.Sprintf("%x", sum[:6])
}

func IsKeyInRange(key string, s string, p string) bool {
	////s -> self
	//s := c.ctx.Value("hostname").(string)
	////p -> predecessor
	//p := c.ctx.Value("predecessor").(string)

	if GenerateSHA1(s) == GenerateSHA1(key) {
		return true
	}
	if GenerateSHA1(p) == GenerateSHA1(s) {
		return true
	}
	if GenerateSHA1(p) < GenerateSHA1(s) {
		return GenerateSHA1(p) < GenerateSHA1(key) && GenerateSHA1(key) <= GenerateSHA1(s)
	} else {
		if GenerateSHA1(key) > GenerateSHA1(p) {
			return true
		}
		if GenerateSHA1(key) <= GenerateSHA1(s) {
			return true
		}
	}
	return false

}

func GenerateSecurePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))
	for i := range password {
		index, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("error generating random index: %v", err)
		}
		password[i] = charset[index.Int64()]
	}

	return string(password), nil
}
