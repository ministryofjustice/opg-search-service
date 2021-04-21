package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"opg-search-service/response"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type authorisationError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type HashedEmail struct{}

type Cacheable interface {
	GetSecretString(key string) (string, error)
}

func JwtVerify(secretsCache Cacheable) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			jwtSecret, jwtErr := secretsCache.GetSecretString("jwt-key")
			if jwtErr != nil {
				log.Println("Error in fetching JWT secret from cache:", jwtErr.Error())
				response.WriteJSONError(rw, "missing_secret_key", jwtErr.Error(), http.StatusInternalServerError)
				return
			}

			header := r.Header.Get("Authorization")

			token, authErr := verifyToken(header, jwtSecret)

			if authErr != nil {
				log.Println("Error in token verification :", authErr.Description)
				response.WriteJSONError(rw, "Authorisation Error", authErr.Error, http.StatusUnauthorized)
			} else {
				claims := token.Claims.(jwt.MapClaims)
				email := claims["session-data"].(string)
				salt, saltErr := secretsCache.GetSecretString("user-hash-salt")
				if saltErr != nil {
					log.Println("Error in fetching hash salt from cache:", saltErr.Error())
					response.WriteJSONError(rw, "missing_secret_salt", saltErr.Error(), http.StatusInternalServerError)
					return
				}
				hashedEmail := hashEmail(email, salt)
				log.Println("JWT Token is valid for user ", hashedEmail)

				ctx := context.WithValue(r.Context(), HashedEmail{}, hashedEmail)
				next.ServeHTTP(rw, r.WithContext(ctx))
			}
		})
	}
}

func verifyToken(header string, secret string) (*jwt.Token, *authorisationError) {
	if header == "" {
		return nil, &authorisationError{Error: "missing_token", Description: "missing authentication token"}
	}

	header = strings.Split(header, "Bearer ")[1]

	token, parseErr := jwt.Parse(header, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if parseErr != nil {
		return nil, &authorisationError{Error: "error_with_token", Description: parseErr.Error()}
	}

	if !token.Valid {
		return nil, &authorisationError{Error: "error_with_token", Description: "invalid authentication token"}
	}

	return token, nil
}

func hashEmail(email string, salt string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(salt + email))
	return hex.EncodeToString(hash.Sum(nil))
}
