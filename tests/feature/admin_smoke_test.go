package feature_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goravel/app/models"
	"goravel/tests"
)

const (
	smokeAdminUsername = "smoke_admin"
	smokeAdminPassword = "SmokeAdmin123!"
)

func ensureSmokeAdmin(t *testing.T) {
	t.Helper()

	exists, err := facades.Orm().Query().Model(&models.Admin{}).Where("username", smokeAdminUsername).Exists()
	require.NoError(t, err)
	if exists {
		return
	}

	hashed, err := facades.Hash().Make(smokeAdminPassword)
	require.NoError(t, err)

	admin := models.Admin{
		Username: smokeAdminUsername,
		Password: hashed,
		Nickname: "Smoke Admin",
		Status:   1,
	}
	require.NoError(t, facades.Orm().Query().Create(&admin))
}

func loginSmokeAdmin(t *testing.T) string {
	t.Helper()
	ensureSmokeAdmin(t)

	body := fmt.Sprintf(`{"username":%q,"password":%q}`, smokeAdminUsername, smokeAdminPassword)
	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/login", strings.NewReader(body))
	require.NoError(t, err)
	resp.AssertOk()

	content, err := resp.Content()
	require.NoError(t, err)

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(content), &payload))
	require.Equal(t, 200, payload.Code)
	require.NotEmpty(t, payload.Data.Token, "login should return token")
	return payload.Data.Token
}

func TestAdminLoginSuccess(t *testing.T) {
	token := loginSmokeAdmin(t)
	assert.NotEmpty(t, token)
}

func TestAdminInfoWithToken(t *testing.T) {
	token := loginSmokeAdmin(t)

	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Get("/api/admin/info")
	require.NoError(t, err)
	resp.AssertOk()

	content, err := resp.Content()
	require.NoError(t, err)

	var payload struct {
		Code int            `json:"code"`
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(content), &payload))
	require.Equal(t, 200, payload.Code)
	require.NotEmpty(t, payload.Data)
}

func TestAdminMenusTreeWithToken(t *testing.T) {
	token := loginSmokeAdmin(t)

	testCase := tests.TestCase{}
	resp, err := testCase.Http(t).
		WithHeader("Authorization", "Bearer "+token).
		Get("/api/admin/menus/tree")
	require.NoError(t, err)
	resp.AssertOk()

	content, err := resp.Content()
	require.NoError(t, err)

	var payload struct {
		Code int `json:"code"`
	}
	require.NoError(t, json.Unmarshal([]byte(content), &payload))
	require.Equal(t, 200, payload.Code)
}
