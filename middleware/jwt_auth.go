package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
				rw.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(rw).Encode(&authorisationError{
					Error:			"missing_secret_key",
					Description: 	jwtErr.Error(),
				}); err != nil {
					log.Println("handler/middleware failed to write response:", err)
				}
				return
			}

			header := r.Header.Get("Authorization")

			token, authErr := verifyToken(header, jwtSecret)

			if authErr != nil {
				rw.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(rw).Encode(authErr); err != nil {
					log.Println("handler/middleware failed to write response:", err)
				}
			} else {
				claims := token.Claims.(jwt.MapClaims)
				email := claims["session-data"].(string)
				salt, saltErr := secretsCache.GetSecretString("user-hash-salt")
				if saltErr != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					if err := json.NewEncoder(rw).Encode(&authorisationError{
						Error:			"missing_secret_salt",
						Description: 	saltErr.Error(),
					}); err != nil {
						log.Println("handler/middleware failed to write response:", err)
					}
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
