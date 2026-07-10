package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	appfacades "goravel/app/facades"
	"sync"
	"time"

	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
	"goravel/app/utils/traceid"
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
	ctx context.Context
}

var (
	personalAccessTokensColumnsOnce sync.Once
	hasPATBrowserColumn             bool
	hasPATIPColumn                  bool
	hasPATOSColumn                  bool
	hasPATSessionIDColumn           bool
)

func NewTokenServiceImpl(ctx context.Context) *TokenServiceImpl {
	return &TokenServiceImpl{
		ctx: ctx}
}

func loadPersonalAccessTokenColumns() {
	personalAccessTokensColumnsOnce.Do(func() {
		if !facades.Schema().HasTable("personal_access_tokens") {
			return
		}
		hasPATBrowserColumn = facades.Schema().HasColumn("personal_access_tokens", "browser")
		hasPATIPColumn = facades.Schema().HasColumn("personal_access_tokens", "ip")
		hasPATOSColumn = facades.Schema().HasColumn("personal_access_tokens", "os")
		hasPATSessionIDColumn = facades.Schema().HasColumn("personal_access_tokens", "session_id")
	})
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
	loadPersonalAccessTokenColumns()
	payload := map[string]any{
		"tokenable_type": tokenableType,
		"tokenable_id":   tokenableID,
		"name":           name,
		"token":          tokenHash,
		"expires_at":     expiresAt,
		"last_used_at":   &now,
	}
	if hasPATBrowserColumn {
		payload["browser"] = browser
	}
	if hasPATIPColumn {
		payload["ip"] = ip
	}
	if hasPATOSColumn {
		payload["os"] = os
	}
	if hasPATSessionIDColumn {
		payload["session_id"] = sessionID
	}

	if err := appfacades.OrmQuery(s.ctx).Table("personal_access_tokens").Create(payload); err != nil {
		return "", nil, err
	}

	var accessToken models.PersonalAccessToken
	if err := appfacades.OrmQuery(s.ctx).Where("token", tokenHash).First(&accessToken); err != nil {
		return plainToken, nil, nil
	}

	return plainToken, &accessToken, nil
}

// FindToken 根据token值查找token记录
func (s *TokenServiceImpl) FindToken(token string) (*models.PersonalAccessToken, error) {
	if token == "" {
		return nil, apperrors.ErrInvalidArgument.WithMessage("token is empty")
	}

	tokenHash := s.hashToken(token)
	var accessToken models.PersonalAccessToken
	if err := appfacades.OrmQuery(s.ctx).Where("token", tokenHash).FirstOrFail(&accessToken); err != nil {
		// 记录调试信息
		// facades.Log().Debugf("TokenService: FindToken failed, hash: %s, error: %v", tokenHash[:min(20, len(tokenHash))], err)
		return nil, err
	}

	// 检查是否过期
	if accessToken.ExpiresAt != nil && accessToken.ExpiresAt.Before(time.Now()) {
		if _, err := appfacades.OrmQuery(s.ctx).Delete(&accessToken); err != nil {
			// 使用 traceid.EnsureContext 确保有 trace_id，即使使用 context.Background()
			// 这样可以保证日志的可追踪性
			ctx, _ := traceid.EnsureContext(context.Background())
			errorlog.Record(ctx, "token", "Failed to delete expired token", map[string]any{
				"token_id":     accessToken.ID,
				"tokenable_id": accessToken.TokenableID,
				"expires_at":   accessToken.ExpiresAt,
				"error":        err.Error(),
			}, "Failed to delete expired token (ID: %d): %v", accessToken.ID, err)
		}

		return nil, apperrors.ErrInvalidArgument.WithMessage("token expired")
	}

	return &accessToken, nil
}

// DeleteToken 删除token
func (s *TokenServiceImpl) DeleteToken(token string) error {
	tokenHash := s.hashToken(token)
	_, err := appfacades.OrmQuery(s.ctx).Where("token", tokenHash).Delete(&models.PersonalAccessToken{})
	return err
}

// DeleteTokensByUser 删除用户的所有token
func (s *TokenServiceImpl) DeleteTokensByUser(tokenableType string, tokenableID uint) error {
	_, err := appfacades.OrmQuery(s.ctx).
		Where("tokenable_type", tokenableType).
		Where("tokenable_id", tokenableID).
		Delete(&models.PersonalAccessToken{})
	return err
}

// GetTokensByUser 获取用户的所有token
func (s *TokenServiceImpl) GetTokensByUser(tokenableType string, tokenableID uint) ([]models.PersonalAccessToken, error) {
	var tokens []models.PersonalAccessToken
	err := appfacades.OrmQuery(s.ctx).
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
	_, err := appfacades.OrmQuery(s.ctx).
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
