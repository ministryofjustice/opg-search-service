package middleware

import (
	"net/http"
)

func ContentType() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(rw, r)
		})
	}
}
