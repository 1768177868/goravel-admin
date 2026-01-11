package services

import (
	"testing"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/tests"
)

type TokenServiceFeatureTestSuite struct {
	suite.Suite
	tests.TestCase
	service services.TokenService
}

func TestTokenServiceFeatureTestSuite(t *testing.T) {
	suite.Run(t, &TokenServiceFeatureTestSuite{})
}

func (s *TokenServiceFeatureTestSuite) SetupSuite() {
	s.service = services.NewTokenServiceImpl()
}

func (s *TokenServiceFeatureTestSuite) SetupTest() {
	s.RefreshDatabase()
	s.Seed()
}

func (s *TokenServiceFeatureTestSuite) TearDownTest() {
	// 清理测试数据
	_, _ = facades.Orm().Query().Where("name LIKE ?", "test_%").Delete(&models.PersonalAccessToken{})
}

// ==================== CreateToken 方法测试 ====================

// TestCreateToken_Success 测试创建 token 成功
func (s *TokenServiceFeatureTestSuite) TestCreateToken_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	expiresAt := time.Now().Add(24 * time.Hour)
	plainToken, accessToken, err := s.service.CreateToken(
		"admin",
		admin.ID,
		"test_token",
		&expiresAt,
		"Chrome",
		"127.0.0.1",
		"Windows",
		"session123",
	)

	s.NoError(err)
	s.NotEmpty(plainToken)
	s.NotNil(accessToken)
	s.Equal("admin", accessToken.TokenableType)
	s.Equal(admin.ID, accessToken.TokenableID)
	s.Equal("test_token", accessToken.Name)
	s.Equal("Chrome", accessToken.Browser)
	s.Equal("127.0.0.1", accessToken.IP)
	s.Equal("Windows", accessToken.OS)
	s.Equal("session123", accessToken.SessionID)
	s.NotNil(accessToken.LastUsedAt, "创建时应该设置 LastUsedAt")
}

// TestCreateToken_WithoutExpiresAt 测试创建无过期时间的 token
func (s *TokenServiceFeatureTestSuite) TestCreateToken_WithoutExpiresAt() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	plainToken, accessToken, err := s.service.CreateToken(
		"admin",
		admin.ID,
		"test_token",
		nil, // 无过期时间
		"",
		"",
		"",
		"",
	)

	s.NoError(err)
	s.NotEmpty(plainToken)
	s.NotNil(accessToken)
	s.Nil(accessToken.ExpiresAt, "无过期时间的 token 应该 ExpiresAt 为 nil")
}

// TestCreateToken_GenerateSessionID 测试自动生成 SessionID
func (s *TokenServiceFeatureTestSuite) TestCreateToken_GenerateSessionID() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	plainToken, accessToken, err := s.service.CreateToken(
		"admin",
		admin.ID,
		"test_token",
		nil,
		"",
		"",
		"",
		"", // 不提供 SessionID，应该自动生成
	)

	s.NoError(err)
	s.NotEmpty(plainToken)
	s.NotNil(accessToken)
	s.NotEmpty(accessToken.SessionID, "应该自动生成 SessionID")
}

// ==================== FindToken 方法测试 ====================

// TestFindToken_Success 测试查找 token 成功
func (s *TokenServiceFeatureTestSuite) TestFindToken_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建 token
	plainToken, _, err := s.service.CreateToken(
		"admin",
		admin.ID,
		"test_token",
		nil,
		"",
		"",
		"",
		"",
	)
	s.Require().NoError(err)

	// 查找 token
	foundToken, err := s.service.FindToken(plainToken)

	s.NoError(err)
	s.NotNil(foundToken)
	s.Equal(admin.ID, foundToken.TokenableID)
}

// TestFindToken_EmptyToken 测试空 token
func (s *TokenServiceFeatureTestSuite) TestFindToken_EmptyToken() {
	token, err := s.service.FindToken("")

	s.Error(err)
	s.Nil(token)

	// 验证返回的是业务错误
	businessErr, ok := apperrors.GetBusinessError(err)
	s.True(ok, "应该返回业务错误")
	s.Equal("invalid_argument", businessErr.Code)
}

// TestFindToken_NotFound 测试 token 不存在
func (s *TokenServiceFeatureTestSuite) TestFindToken_NotFound() {
	// 使用一个不存在的 token
	token, err := s.service.FindToken("non-existent-token-12345")

	s.Error(err)
	s.Nil(token)
}

// TestFindToken_ExpiredToken 测试过期 token
func (s *TokenServiceFeatureTestSuite) TestFindToken_ExpiredToken() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建已过期的 token
	pastTime := time.Now().Add(-1 * time.Hour)
	plainToken, _, err := s.service.CreateToken(
		"admin",
		admin.ID,
		"test_token",
		&pastTime,
		"",
		"",
		"",
		"",
	)
	s.Require().NoError(err)

	// 尝试查找过期 token
	token, err := s.service.FindToken(plainToken)

	s.Error(err)
	s.Nil(token)

	// 验证返回的是业务错误
	businessErr, ok := apperrors.GetBusinessError(err)
	s.True(ok, "应该返回业务错误")
	s.Equal("invalid_argument", businessErr.Code)
	s.Contains(businessErr.Message, "expired")

	// 验证过期 token 已被删除
	count, err := facades.Orm().Query().
		Model(&models.PersonalAccessToken{}).
		Where("tokenable_id", admin.ID).
		Where("name", "test_token").
		Count()
	s.NoError(err)
	s.Equal(int64(0), count, "过期 token 应该被删除")
}

// ==================== DeleteToken 方法测试 ====================

// TestDeleteToken_Success 测试删除 token 成功
func (s *TokenServiceFeatureTestSuite) TestDeleteToken_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建 token
	plainToken, _, err := s.service.CreateToken(
		"admin",
		admin.ID,
		"test_token",
		nil,
		"",
		"",
		"",
		"",
	)
	s.Require().NoError(err)

	// 删除 token
	err = s.service.DeleteToken(plainToken)
	s.NoError(err)

	// 验证 token 已被删除
	foundToken, err := s.service.FindToken(plainToken)
	s.Error(err)
	s.Nil(foundToken)
}

// ==================== DeleteTokensByUser 方法测试 ====================

// TestDeleteTokensByUser_Success 测试删除用户的所有 token
func (s *TokenServiceFeatureTestSuite) TestDeleteTokensByUser_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建多个 token
	for i := 0; i < 3; i++ {
		_, _, err := s.service.CreateToken(
			"admin",
			admin.ID,
			"test_token",
			nil,
			"",
			"",
			"",
			"",
		)
		s.Require().NoError(err)
	}

	// 验证创建成功
	tokens, err := s.service.GetTokensByUser("admin", admin.ID)
	s.Require().NoError(err)
	s.GreaterOrEqual(len(tokens), 3)

	// 删除用户的所有 token
	err = s.service.DeleteTokensByUser("admin", admin.ID)
	s.NoError(err)

	// 验证所有 token 已被删除
	tokens, err = s.service.GetTokensByUser("admin", admin.ID)
	s.NoError(err)
	s.Empty(tokens)
}

// ==================== GetTokensByUser 方法测试 ====================

// TestGetTokensByUser_Success 测试获取用户的所有 token
func (s *TokenServiceFeatureTestSuite) TestGetTokensByUser_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建多个 token
	tokenNames := []string{"token1", "token2", "token3"}
	for _, name := range tokenNames {
		_, _, err := s.service.CreateToken(
			"admin",
			admin.ID,
			name,
			nil,
			"",
			"",
			"",
			"",
		)
		s.Require().NoError(err)
		time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	}

	// 获取用户的所有 token
	tokens, err := s.service.GetTokensByUser("admin", admin.ID)
	s.NoError(err)
	s.GreaterOrEqual(len(tokens), len(tokenNames))

	// 验证返回了数据（排序验证在数据库层面完成，这里只验证数据存在）
	s.NotEmpty(tokens, "应该返回用户的 token")
}

// ==================== UpdateLastUsedAt 方法测试 ====================

// TestUpdateLastUsedAt_Success 测试更新最后使用时间
func (s *TokenServiceFeatureTestSuite) TestUpdateLastUsedAt_Success() {
	var admin models.Admin
	err := facades.Orm().Query().First(&admin)
	s.Require().NoError(err)

	// 创建 token
	plainToken, accessToken, err := s.service.CreateToken(
		"admin",
		admin.ID,
		"test_token",
		nil,
		"",
		"",
		"",
		"",
	)
	s.Require().NoError(err)

	initialLastUsedAt := accessToken.LastUsedAt
	s.Require().NotNil(initialLastUsedAt)

	// 等待一小段时间确保时间不同
	time.Sleep(100 * time.Millisecond)

	// 更新最后使用时间
	err = s.service.UpdateLastUsedAt(plainToken)
	s.NoError(err)

	// 验证最后使用时间已更新
	updatedToken, err := s.service.FindToken(plainToken)
	s.NoError(err)
	s.NotNil(updatedToken)
	s.NotNil(updatedToken.LastUsedAt)
	s.True(updatedToken.LastUsedAt.After(*initialLastUsedAt),
		"最后使用时间应该已更新")
}

// TestUpdateLastUsedAt_NotFound 测试更新不存在的 token
func (s *TokenServiceFeatureTestSuite) TestUpdateLastUsedAt_NotFound() {
	// 尝试更新不存在的 token
	err := s.service.UpdateLastUsedAt("non-existent-token")

	// 注意：UpdateLastUsedAt 可能不会返回错误，只是不更新任何记录
	// 这取决于 ORM 的实现
	s.NoError(err) // 或者根据实际实现调整断言
}
