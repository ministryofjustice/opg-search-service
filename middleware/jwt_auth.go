package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type authorisationError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

type HashedEmail struct{}

func JwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		jwtSecret := os.Getenv("JWT_SECRET")
		salt := os.Getenv("USER_HASH_SALT")

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
			hashedEmail := hashEmail(email, salt)
			log.Println("JWT Token is valid for user ", hashedEmail)

			ctx := context.WithValue(r.Context(), HashedEmail{}, hashedEmail)
			next.ServeHTTP(rw, r.WithContext(ctx))
		}
	})
}

func verifyToken(header string, secret string) (*jwt.Token, *authorisationError) {
	if header == "" {
		return nil, &authorisationError{Error: "missing_token", Description: "missing authentication token"}
	}

	header = strings.Split(header, "Bearer ")[1]

	token, err := jwt.Parse(header, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, &authorisationError{Error: "error_with_token", Description: err.Error()}
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
