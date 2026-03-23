package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func loadPrivateKey() (*rsa.PrivateKey, error) {
	keyData := os.Getenv("GITHUB_PRIVATE_KEY")
	// Allow \n literals in the env var to be treated as real newlines
	keyData = strings.ReplaceAll(keyData, `\n`, "\n")

	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
func GenerateGitHubAppJWT() (string, error) {
	privateKey, err := loadPrivateKey()
	if err != nil {
		return "", err
	}

	now := time.Now()

	claims := jwt.MapClaims{
		"iat": now.Unix() - 60,                  // issued 1 min in past
		"exp": now.Add(10 * time.Minute).Unix(), // max 10 minutes
		"iss": os.Getenv("GITHUB_APP_ID"),       // your App ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
