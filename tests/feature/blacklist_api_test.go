package feature

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
	"goravel/tests"
)

type BlacklistApiTestSuite struct {
	suite.Suite
	tests.TestCase
	token string
}

func TestBlacklistApiTestSuite(t *testing.T) {
	suite.Run(t, &BlacklistApiTestSuite{})
}

func (s *BlacklistApiTestSuite) SetupSuite() {
}

func (s *BlacklistApiTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed()
	s.token = s.getAdminToken()
}

func (s *BlacklistApiTestSuite) TearDownTest() {
}

func (s *BlacklistApiTestSuite) getAdminToken() string {
	body := strings.NewReader(`{"username":"admin","password":"admin123"}`)
	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/login", body)

	s.Require().NoError(err)

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	_ = json.Unmarshal([]byte(content), &result)
	data := result["data"].(map[string]any)
	return data["token"].(string)
}

// ==================== 黑名单列表测试 ====================

func (s *BlacklistApiTestSuite) TestBlacklistList_Success() {
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Get("/api/admin/blacklists")

	s.Require().NoError(err)
	resp.AssertSuccessful()

	content, err := resp.Content()
	s.Require().NoError(err)

	var result map[string]any
	err = json.Unmarshal([]byte(content), &result)
	s.Require().NoError(err)

	s.Equal(float64(200), result["code"])
}

// ==================== 黑名单创建测试 ====================

func (s *BlacklistApiTestSuite) TestBlacklistCreate_SingleIP() {
	body := strings.NewReader(`{
		"ip": "192.168.1.100",
		"remark": "测试单个IP黑名单",
		"status": 1
	}`)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/blacklists", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证黑名单已创建
	var blacklist models.Blacklist
	err = facades.Orm().Query().Where("ip", "192.168.1.100").First(&blacklist)
	s.Require().NoError(err)
	s.Equal("192.168.1.100", blacklist.IP)
	s.Equal("测试单个IP黑名单", blacklist.Remark)
}

func (s *BlacklistApiTestSuite) TestBlacklistCreate_CIDR() {
	body := strings.NewReader(`{
		"ip": "10.0.0.0/24",
		"remark": "测试CIDR格式黑名单",
		"status": 1
	}`)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/blacklists", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证黑名单已创建
	var blacklist models.Blacklist
	err = facades.Orm().Query().Where("ip", "10.0.0.0/24").First(&blacklist)
	s.Require().NoError(err)
	s.Equal("10.0.0.0/24", blacklist.IP)
}

func (s *BlacklistApiTestSuite) TestBlacklistCreate_IPRange() {
	body := strings.NewReader(`{
		"ip": "172.16.0.1-172.16.0.100",
		"remark": "测试IP范围黑名单",
		"status": 1
	}`)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/blacklists", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()
}

func (s *BlacklistApiTestSuite) TestBlacklistCreate_InvalidIP() {
	body := strings.NewReader(`{
		"ip": "invalid-ip-address",
		"remark": "无效IP",
		"status": 1
	}`)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Post("/api/admin/blacklists", body)

	s.Require().NoError(err)
	// 应该返回验证错误
	resp.AssertStatus(422)
}

// ==================== 黑名单更新测试 ====================

func (s *BlacklistApiTestSuite) TestBlacklistUpdate_Success() {
	// 先创建一个黑名单
	blacklist := models.Blacklist{
		IP:     "192.168.2.1",
		Remark: "原始备注",
		Status: 1,
	}
	err := facades.Orm().Query().Create(&blacklist)
	s.Require().NoError(err)

	// 更新黑名单
	body := strings.NewReader(`{
		"ip": "192.168.2.2",
		"remark": "更新后的备注",
		"status": 0
	}`)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Put(fmt.Sprintf("/api/admin/blacklists/%d", blacklist.ID), body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证更新
	var updated models.Blacklist
	err = facades.Orm().Query().Find(&updated, blacklist.ID)
	s.Require().NoError(err)
	s.Equal("192.168.2.2", updated.IP)
	s.Equal("更新后的备注", updated.Remark)
	s.Equal(uint8(0), updated.Status)
}

// ==================== 黑名单删除测试 ====================

func (s *BlacklistApiTestSuite) TestBlacklistDelete_Success() {
	// 先创建一个黑名单
	blacklist := models.Blacklist{
		IP:     "192.168.3.1",
		Remark: "待删除",
		Status: 1,
	}
	err := facades.Orm().Query().Create(&blacklist)
	s.Require().NoError(err)

	// 删除黑名单
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		Delete(fmt.Sprintf("/api/admin/blacklists/%d", blacklist.ID))

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证已删除
	var count int64
	facades.Orm().Query().Model(&models.Blacklist{}).Where("id", blacklist.ID).Count(&count)
	s.Equal(int64(0), count)
}

// ==================== 黑名单批量删除测试 ====================

func (s *BlacklistApiTestSuite) TestBlacklistBatchDelete_Success() {
	// 创建多个黑名单
	var ids []uint
	for i := 1; i <= 3; i++ {
		blacklist := models.Blacklist{
			IP:     fmt.Sprintf("192.168.10.%d", i),
			Remark: "批量删除测试",
			Status: 1,
		}
		err := facades.Orm().Query().Create(&blacklist)
		s.Require().NoError(err)
		ids = append(ids, blacklist.ID)
	}

	// 批量删除
	body := strings.NewReader(fmt.Sprintf(`{"ids":[%d,%d,%d]}`, ids[0], ids[1], ids[2]))
	resp, err := s.Http(s.T()).
		WithHeader("Authorization", "Bearer "+s.token).
		WithHeader("Content-Type", "application/json").
		Delete("/api/admin/blacklists/batch", body)

	s.Require().NoError(err)
	resp.AssertSuccessful()

	// 验证已删除
	var count int64
	facades.Orm().Query().Model(&models.Blacklist{}).Where("id IN ?", ids).Count(&count)
	s.Equal(int64(0), count)
}
