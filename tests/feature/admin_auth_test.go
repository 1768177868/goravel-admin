package feature_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"goravel/tests"
)

func TestAdminLoginValidation(t *testing.T) {
	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/login", strings.NewReader(`{}`))
	assert.NoError(t, err)
	resp.AssertBadRequest()
}

func TestAdminProtectedRouteUnauthorized(t *testing.T) {
	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).Get("/api/admin/info")
	assert.NoError(t, err)
	resp.AssertUnauthorized()
}
