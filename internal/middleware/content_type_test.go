package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestContentType(t *testing.T) {
    req, err := http.NewRequest("GET", "/", nil)
    if err != nil {
        t.Fatal(err)
    }

    testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

    rw := httptest.NewRecorder()

    handler := ContentType()(testHandler)
    handler.ServeHTTP(rw, req)
    assert.Equal(
        t,
        "application/json",
        rw.Header().Get("Content-Type"),
        "Content-Type header should be application/json",
    )

}
