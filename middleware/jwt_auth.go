package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"opg-search-service/response"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

type authorisationError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type HashedEmail struct{}

type Cacheable interface {
	GetSecretString(key string) (string, error)
}

func JwtVerify(secretsCache Cacheable, logger *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			jwtSecret, jwtErr := secretsCache.GetSecretString("jwt-key")
			if jwtErr != nil {
				logger.Println("Error in fetching JWT secret from cache:", jwtErr.Error())
				response.WriteJSONError(rw, "missing_secret_key", jwtErr.Error(), http.StatusInternalServerError)
				return
			}

			header := r.Header.Get("Authorization")

			token, verifyErr := verifyToken(header, jwtSecret)

			if verifyErr != nil {
				logger.Println("Error in token verification :", verifyErr.Error())
				response.WriteJSONError(rw, "Authorisation Error", verifyErr.Error(), http.StatusUnauthorized)
			} else {
				claims := token.Claims.(jwt.MapClaims)
				email := claims["session-data"].(string)
				salt, saltErr := secretsCache.GetSecretString("user-hash-salt")
				if saltErr != nil {
					logger.Println("Error in fetching hash salt from cache:", saltErr.Error())
					response.WriteJSONError(rw, "missing_secret_salt", saltErr.Error(), http.StatusInternalServerError)
					return
				}
				hashedEmail := hashEmail(email, salt)
				logger.Println("JWT Token is valid for user ", hashedEmail)

				ctx := context.WithValue(r.Context(), HashedEmail{}, hashedEmail)
				next.ServeHTTP(rw, r.WithContext(ctx))
			}
		})
	}
}

func verifyToken(header string, secret string) (*jwt.Token, error) {
	if header == "" {
		return nil, errors.New("missing authentication token")
	}

	header = strings.Split(header, "Bearer ")[1]

	token, parseErr := jwt.Parse(header, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if parseErr != nil {
		return nil, errors.New(parseErr.Error())
	}

	return token, nil
}

func hashEmail(email string, salt string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(salt + email))
	return hex.EncodeToString(hash.Sum(nil))
}
