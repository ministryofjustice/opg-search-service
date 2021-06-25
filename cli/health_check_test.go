package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthCheck(t *testing.T) {
	l, _ := test.NewNullLogger()
	hc := NewHealthCheck(l)

	assert.IsType(t, new(healthCheck), hc)
	assert.Same(t, l, hc.logger)
	assert.Nil(t, hc.shouldRun)
	assert.Equal(t, "http://localhost:8000"+os.Getenv("PATH_PREFIX")+"/health-check", hc.checkUrl)
	assert.IsType(t, os.Exit, hc.exit)
}

func TestHealthCheck_DefineFlags(t *testing.T) {
	l, _ := test.NewNullLogger()
	hc := NewHealthCheck(l)
	assert.Nil(t, hc.shouldRun)
	hc.DefineFlags()
	assert.False(t, *hc.shouldRun)
}

func TestHealthCheck_ShouldRun(t *testing.T) {
	tests := []struct {
		scenario  string
		shouldRun bool
	}{
		{
			scenario:  "Command should run",
			shouldRun: true,
		},
		{
			scenario:  "Command should not run",
			shouldRun: false,
		},
	}
	for _, test := range tests {
		hc := &healthCheck{
			shouldRun: &test.shouldRun,
		}
		assert.Equal(t, test.shouldRun, hc.ShouldRun(), test.scenario)
	}
}

func TestHealthCheck_Run(t *testing.T) {
	tests := []struct {
		scenario     string
		responseCode int
		wantInLog    string
		wantExitCode int
	}{
		{
			scenario:     "PASS",
			responseCode: 200,
			wantInLog:    "OK",
			wantExitCode: 0,
		},
		{
			scenario:     "FAIL",
			responseCode: 500,
			wantInLog:    "FAIL",
			wantExitCode: 1,
		},
	}
	for _, tc := range tests {
		l, hook := test.NewNullLogger()

		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tc.responseCode)
		}))

		exitCode := 666

		hc := &healthCheck{
			logger:   l,
			checkUrl: s.URL,
			exit: func(code int) {
				exitCode = code
			},
		}
		hc.Run()

		assert.Contains(t, hook.LastEntry().Message, tc.wantInLog, tc.scenario)
		assert.Equal(t, tc.wantExitCode, exitCode, tc.scenario)
	}
}
