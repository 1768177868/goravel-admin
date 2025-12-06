package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/models"
)

type TokenService interface {
	// CreateToken 创建token并存入数据库
	CreateToken(tokenableType string, tokenableID uint, name string, expiresAt *time.Time, browser, ip, os, sessionID string) (string, *models.PersonalAccessToken, error)
	// FindToken 根据token值查找token记录
	FindToken(token string) (*models.PersonalAccessToken, error)
	// DeleteToken 删除token
	DeleteToken(token string) error
	// DeleteTokensByUser 删除用户的所有token
	DeleteTokensByUser(tokenableType string, tokenableID uint) error
	// GetTokensByUser 获取用户的所有token
	GetTokensByUser(tokenableType string, tokenableID uint) ([]models.PersonalAccessToken, error)
	// UpdateLastUsedAt 更新最后使用时间
	UpdateLastUsedAt(token string) error
}

type TokenServiceImpl struct {
}

func NewTokenServiceImpl() *TokenServiceImpl {
	return &TokenServiceImpl{}
}

// CreateToken 创建token并存入数据库
func (s *TokenServiceImpl) CreateToken(tokenableType string, tokenableID uint, name string, expiresAt *time.Time, browser, ip, os, sessionID string) (string, *models.PersonalAccessToken, error) {
	// 生成随机token（类似Laravel Sanctum）
	plainToken := s.generateRandomToken()
	tokenHash := s.hashToken(plainToken)

	// 如果没有提供sessionID，使用tokenHash的前16位作为sessionID
	if sessionID == "" {
		if len(tokenHash) >= 16 {
			sessionID = tokenHash[:16]
		} else {
			sessionID = tokenHash
		}
	}

	// 创建token记录，立即设置last_used_at为当前时间
	now := time.Now()
	accessToken := &models.PersonalAccessToken{
		TokenableType: tokenableType,
		TokenableID:   tokenableID,
		Name:          name,
		Token:         tokenHash,
		ExpiresAt:     expiresAt,
		LastUsedAt:    &now, // 登录时立即设置最后使用时间
		Browser:       browser,
		IP:            ip,
		OS:            os,
		SessionID:     sessionID,
	}

	if err := facades.Orm().Query().Create(accessToken); err != nil {
		return "", nil, err
	}

	return plainToken, accessToken, nil
}

// FindToken 根据token值查找token记录
// 
// 参数:
//   - token: token 字符串
//
// 返回:
//   - *models.PersonalAccessToken: token 记录
//   - error: 错误信息
func (s *TokenServiceImpl) FindToken(token string) (*models.PersonalAccessToken, error) {
	if token == "" {
		return nil, apperrors.ErrInvalidArgument.WithMessage("token is empty")
	}

	tokenHash := s.hashToken(token)
	var accessToken models.PersonalAccessToken
	if err := facades.Orm().Query().Where("token", tokenHash).First(&accessToken); err != nil {
		// 记录调试信息
		// facades.Log().Debugf("TokenService: FindToken failed, hash: %s, error: %v", tokenHash[:min(20, len(tokenHash))], err)
		return nil, err
	}

	// 检查是否过期
	if accessToken.ExpiresAt != nil && accessToken.ExpiresAt.Before(time.Now()) {
		// token已过期，删除它
		_, _ = facades.Orm().Query().Delete(&accessToken)
		// 返回错误，表示token已过期
		return nil, apperrors.ErrInvalidArgument.WithMessage("token expired")
	}

	return &accessToken, nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DeleteToken 删除token
func (s *TokenServiceImpl) DeleteToken(token string) error {
	tokenHash := s.hashToken(token)
	_, err := facades.Orm().Query().Where("token", tokenHash).Delete(&models.PersonalAccessToken{})
	return err
}

// DeleteTokensByUser 删除用户的所有token
func (s *TokenServiceImpl) DeleteTokensByUser(tokenableType string, tokenableID uint) error {
	_, err := facades.Orm().Query().
		Where("tokenable_type", tokenableType).
		Where("tokenable_id", tokenableID).
		Delete(&models.PersonalAccessToken{})
	return err
}

// GetTokensByUser 获取用户的所有token
func (s *TokenServiceImpl) GetTokensByUser(tokenableType string, tokenableID uint) ([]models.PersonalAccessToken, error) {
	var tokens []models.PersonalAccessToken
	err := facades.Orm().Query().
		Where("tokenable_type", tokenableType).
		Where("tokenable_id", tokenableID).
		Order("created_at desc").
		Find(&tokens)
	return tokens, err
}

// UpdateLastUsedAt 更新最后使用时间
func (s *TokenServiceImpl) UpdateLastUsedAt(token string) error {
	tokenHash := s.hashToken(token)
	now := time.Now()
	_, err := facades.Orm().Query().
		Model(&models.PersonalAccessToken{}).
		Where("token", tokenHash).
		Update("last_used_at", now)
	return err
}

// generateRandomToken 生成随机token（40个字符）
func (s *TokenServiceImpl) generateRandomToken() string {
	// 生成40个随机字节
	b := make([]byte, 40)
	_, _ = rand.Read(b)
	// 转换为十六进制字符串（80个字符），然后取前40个字符
	return hex.EncodeToString(b)[:40]
}

// hashToken 对token进行SHA256哈希
func (s *TokenServiceImpl) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
