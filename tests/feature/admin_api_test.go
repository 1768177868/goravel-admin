package feature

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
	"goravel/tests"
)

type AdminApiTestSuite struct {
	suite.Suite
	tests.TestCase
	token string
}

func TestAdminApiTestSuite(t *testing.T) {
	suite.Run(t, &AdminApiTestSuite{})
}

func (s *AdminApiTestSuite) SetupSuite() {
}

func (s *AdminApiTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed()
	s.token = s.getAdminToken()
}

func (s *AdminApiTestSuite) TearDownTest() {
}

// getAdminToken 获取管理员登录 token
func (s *AdminApiTestSuite) getAdminToken() string {
	body := strings.NewReader(`{"username":"admin","password":"admin123"}`)
	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/login", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	err = json.Unmarshal([]byte(content), &result)
	s.Require().NoError(err)

	data, ok := result["data"].(map[string]any)
	s.Require().True(ok)

	token, ok := data["token"].(string)
	s.Require().True(ok)

	return token
}

// ==================== 登录接口测试 ====================

func (s *AdminApiTestSuite) TestLogin_Success() {
	body := strings.NewReader(`{"username":"admin","password":"admin123"}`)
	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/login", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	err = json.Unmarshal([]byte(content), &result)
	s.Require().NoError(err)

	s.Equal(float64(200), result["code"])
	s.NotNil(result["data"])
}

func (s *AdminApiTestSuite) TestLogin_WrongPassword() {
	body := strings.NewReader(`{"username":"admin","password":"wrongpassword"}`)
	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/login", body)

	s.Require().NoError(err)
	// 登录失败应该返回 401
	resp.AssertUnauthorized()
}

func (s *AdminApiTestSuite) TestLogin_UserNotFound() {
	body := strings.NewReader(`{"username":"nonexistent","password":"test123"}`)
	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/login", body)

	s.Require().NoError(err)
	resp.AssertUnauthorized()
}

// ==================== 管理员信息接口测试 ====================

func (s *AdminApiTestSuite) TestGetAdminInfo_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/info")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	err = json.Unmarshal([]byte(content), &result)
	s.Require().NoError(err)

	s.Equal(float64(200), result["code"])
	data, ok := result["data"].(map[string]any)
	s.Require().True(ok)
	s.NotNil(data["username"])
}

func (s *AdminApiTestSuite) TestGetAdminInfo_Unauthorized() {
	resp, err := s.Http(s.T()).Get("/api/admin/info")

	s.Require().NoError(err)
	resp.AssertUnauthorized()
}

// ==================== 管理员列表接口测试 ====================

func (s *AdminApiTestSuite) TestAdminList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/admins")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	err = json.Unmarshal([]byte(content), &result)
	s.Require().NoError(err)

	s.Equal(float64(200), result["code"])
	data, ok := result["data"].(map[string]any)
	s.Require().True(ok)

	// 应该有列表
	list, ok := data["list"].([]any)
	s.Require().True(ok)
	s.Greater(len(list), 0)
}

func (s *AdminApiTestSuite) TestAdminList_WithPagination() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/admins?page=1&page_size=5")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	err = json.Unmarshal([]byte(content), &result)
	s.Require().NoError(err)

	data, ok := result["data"].(map[string]any)
	s.Require().True(ok)

	s.Equal(float64(1), data["page"])
	s.Equal(float64(5), data["page_size"])
}

// ==================== 角色接口测试 ====================

func (s *AdminApiTestSuite) TestRoleList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/roles")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	err = json.Unmarshal([]byte(content), &result)
	s.Require().NoError(err)

	s.Equal(float64(200), result["code"])
}

func (s *AdminApiTestSuite) TestRoleCreate_Success() {
	body := strings.NewReader(`{
		"name": "测试角色",
		"slug": "test-role",
		"description": "测试角色描述",
		"status": 1,
		"sort": 100
	}`)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/roles", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证角色已创建
	var role models.Role
	err = facades.Orm().Query().Where("slug", "test-role").First(&role)
	s.Require().NoError(err)
	s.Equal("测试角色", role.Name)
}

// ==================== 菜单接口测试 ====================

func (s *AdminApiTestSuite) TestMenuList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/menus")

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

// ==================== 部门接口测试 ====================

func (s *AdminApiTestSuite) TestDepartmentList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/departments")

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *AdminApiTestSuite) TestDepartmentCreate_Success() {
	body := strings.NewReader(`{
		"name": "测试部门",
		"parent_id": 0,
		"status": 1,
		"sort": 100
	}`)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/departments", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证部门已创建
	var dept models.Department
	err = facades.Orm().Query().Where("name", "测试部门").First(&dept)
	s.Require().NoError(err)
	s.Equal("测试部门", dept.Name)
}

// ==================== 字典接口测试 ====================

func (s *AdminApiTestSuite) TestDictionaryList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/dictionaries")

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

// ==================== 日志接口测试 ====================

func (s *AdminApiTestSuite) TestOperationLogList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/logs/operation")

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *AdminApiTestSuite) TestLoginLogList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/logs/login")

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *AdminApiTestSuite) TestSystemLogList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/logs/system")

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

// ==================== 仪表盘接口测试 ====================

func (s *AdminApiTestSuite) TestDashboard_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/dashboard/stats")

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

// ==================== 退出登录测试 ====================

func (s *AdminApiTestSuite) TestLogout_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Post("/api/admin/logout", nil)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证 token 已失效
	resp, err = s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/info")

	s.Require().NoError(err)
	resp.AssertUnauthorized()
}
