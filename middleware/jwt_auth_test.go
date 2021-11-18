package middleware

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSecretsCache struct {
	mock.Mock
}

type mockValue struct {
	v string
	e error
}

func (c *mockSecretsCache) GetSecretString(key string) (string, error) {
	args := c.Called(key)
	return args.String(0), args.Error(1)
}

func makeToken(expired bool) string {
	exp := time.Now().AddDate(0, 1, 0).Unix()
	if expired {
		exp = time.Now().AddDate(0, -1, 0).Unix()
	}
	iat := time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix()
	signing := jwt.SigningMethodHS256
	tokenString := buildToken(signing, iat, exp)
	return tokenString
}

func makeInvalidToken() string {
	//identical issue and expiry
	exp := time.Now().AddDate(0, -1, 0).Unix()
	iat := time.Now().AddDate(0, -1, 0).Unix()
	signing := jwt.SigningMethodHS256
	tokenString := buildToken(signing, iat, exp)
	return tokenString
}

func buildToken(signing jwt.SigningMethod, iat int64, exp int64) string {
	token := jwt.NewWithClaims(signing, jwt.MapClaims{
		"session-data": "Test.McTestFace@mail.com",
		"iat":          iat,
		"exp":          exp,
	})
	tokenString, err := token.SignedString([]byte("MyTestSecret"))
	if err != nil {
		log.Fatal("Could not make test token")
	}
	return tokenString
}

func TestJwtVerify(t *testing.T) {
	tests := []struct {
		scenario     string
		token        string
		secret       mockValue
		salt         mockValue
		expectedCode int
	}{
		{
			"Valid token",
			makeToken(false),
			mockValue{"MyTestSecret", nil},
			mockValue{"ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0", nil},
			200,
		},
		{
			"Invalid token",
			makeInvalidToken(),
			mockValue{"MyTestSecret", nil},
			mockValue{"ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0", nil},
			401,
		},
		{
			"No token",
			"",
			mockValue{"MyTestSecret", nil},
			mockValue{"ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0", nil},
			401,
		},
		{
			"Wrong signing method",
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OG85AmiExREkrS6tDfTQ2B3WXlrr-wp5AokiRbz3_oB4OxG-W9KcEEbDRcZc0nH3L7LzYptiy1PtAylQGxHTWZXtGz4ht0bAecBgmpdgXMguEIcoqPJ1n3pIWk_dUZegpqx0Lka21H6XxUTxiy8OcaarA8zdnPUnV6AmNP3ecFawIFYdvJB_cm-GvpCSbr8G8y_Mllj8f4x9nBH8pQux89_6gUY618iYv7tuPWBFfEbLxtF2pZS6YC1aSfLQxeNe8djT9YjpvRZA",
			mockValue{"MyTestSecret", nil},
			mockValue{"ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0", nil},
			401,
		},
		{
			"Expired token",
			makeToken(true),
			mockValue{"MyTestSecret", nil},
			mockValue{"ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0", nil},
			401,
		},
		{
			"Cannot fetch JWT secret",
			makeToken(false),
			mockValue{"", errors.New("Missing secret")},
			mockValue{"ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0", nil},
			500,
		},
		{
			"Cannot fetch salt secret",
			makeToken(false),
			mockValue{"MyTestSecret", nil},
			mockValue{"", errors.New("Missing secret")},
			500,
		},
	}

	for _, tc := range tests {
		req, err := http.NewRequest("GET", "/jwt", nil)
		if err != nil {
			t.Fatal(err)
		}

		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		rw := httptest.NewRecorder()

		logger, _ := test.NewNullLogger()

		mockCache := new(mockSecretsCache)
		mockCache.On("GetSecretString", "jwt-key").Return(tc.secret.v, tc.secret.e)
		mockCache.On("GetSecretString", "user-hash-salt").Return(tc.salt.v, tc.salt.e)
		handler := JwtVerify(mockCache, logger)(testHandler)
		handler.ServeHTTP(rw, req)
		res := rw.Result()
		assert.Equal(t, tc.expectedCode, res.StatusCode, tc.scenario)
	}
}

func TestHashEmail(t *testing.T) {
	assert.Equal(
		t,
		"d1a046e6300ea9a75cc4f9eda85e8442c3e9913b8eeb4ed0895896571e479a99",
		hashEmail("Test.McTestFace@mail.com", "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0"),
	)
}
