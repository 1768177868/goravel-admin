package unit

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/suite"
)

// TokenServiceTestSuite Token 服务测试套件
type TokenServiceTestSuite struct {
	suite.Suite
}

func TestTokenServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TokenServiceTestSuite))
}

// SetupTest 每个测试前执行
func (s *TokenServiceTestSuite) SetupTest() {
}

// TearDownTest 每个测试后执行
func (s *TokenServiceTestSuite) TearDownTest() {
}

// hashToken 模拟 TokenService 的哈希函数
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// TestHashToken_NormalToken 测试正常 token 哈希
func (s *TokenServiceTestSuite) TestHashToken_NormalToken() {
	token := "test-token-123"
	result := hashToken(token)

	s.Len(result, 64, "SHA256 应产生 64 个十六进制字符")
}

// TestHashToken_EmptyToken 测试空 token 哈希
func (s *TokenServiceTestSuite) TestHashToken_EmptyToken() {
	token := ""
	result := hashToken(token)

	s.Len(result, 64, "空 token 也应产生 64 个十六进制字符")
}

// TestHashToken_Consistency 测试哈希一致性
func (s *TokenServiceTestSuite) TestHashToken_Consistency() {
	token := "test-token"
	result1 := hashToken(token)
	result2 := hashToken(token)

	s.Equal(result1, result2, "相同输入应产生相同输出")
}

// TestHashToken_Uniqueness 测试不同输入产生不同哈希
func (s *TokenServiceTestSuite) TestHashToken_Uniqueness() {
	result1 := hashToken("token1")
	result2 := hashToken("token2")

	s.NotEqual(result1, result2, "不同输入应产生不同输出")
}

// TestHashToken_TableDriven 表格驱动测试
func (s *TokenServiceTestSuite) TestHashToken_TableDriven() {
	tests := []struct {
		name    string
		token   string
		wantLen int
	}{
		{"正常 token", "test-token-123", 64},
		{"空 token", "", 64},
		{"长 token", "this-is-a-very-long-token-that-should-still-work", 64},
		{"特殊字符", "token@#$%^&*()", 64},
		{"Unicode", "中文token测试", 64},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := hashToken(tt.token)
			s.Len(got, tt.wantLen)
		})
	}
}
