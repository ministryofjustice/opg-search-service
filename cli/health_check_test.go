package cli

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckRun(t *testing.T) {
	tests := []struct {
		scenario     string
		responseCode int
		wantInLog    []string
		wantErr      error
	}{
		{
			scenario:     "PASS",
			responseCode: 200,
			wantInLog:    []string{"OK"},
			wantErr:      nil,
		},
		{
			scenario:     "FAIL",
			responseCode: 500,
			wantInLog:    []string{},
			wantErr:      errors.New("FAIL"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			l, hook := test.NewNullLogger()

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.responseCode)
			}))

			hc := &healthCheck{
				logger:   l,
				checkUrl: s.URL,
			}
			err := hc.Run([]string{})

			assert.Equal(t, tc.wantErr, err)
			for i, message := range tc.wantInLog {
				assert.Contains(t, message, hook.Entries[i].Message)
			}
		})
	}
}
