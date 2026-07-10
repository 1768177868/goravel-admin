package feature_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"goravel/tests"
)

func TestHealthEndpoint(t *testing.T) {
	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).Get("/health")
	assert.NoError(t, err)
	resp.AssertOk()
	resp.AssertJson(map[string]any{"status": "healthy"})
}

func TestAPIUserRegisterValidation(t *testing.T) {
	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).
		WithHeader("Content-Type", "application/json").
		Post("/api/user/register", strings.NewReader(`{}`))
	assert.NoError(t, err)
	resp.AssertBadRequest()
}

func TestAPIUserLoginValidation(t *testing.T) {
	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).
		WithHeader("Content-Type", "application/json").
		Post("/api/user/login", strings.NewReader(`{}`))
	assert.NoError(t, err)
	resp.AssertBadRequest()
}
