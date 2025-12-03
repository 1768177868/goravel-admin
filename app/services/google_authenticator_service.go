package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/goravel/framework/facades"
)

type GoogleAuthenticatorService interface {
	// GenerateSecret 生成密钥
	GenerateSecret(accountName string) (secret string, qrCodeURL string, err error)
	// GenerateQRCodeImage 生成二维码图片（base64）
	GenerateQRCodeImage(accountName, secret string) (string, error)
	// Verify 验证验证码
	Verify(secret, code string) bool
	// IsBound 检查管理员是否绑定了谷歌验证码
	IsBound(adminID uint) (bool, error)
	// GetSecret 获取管理员的密钥（用于绑定确认）
	GetSecret(adminID uint) (string, error)
	// Bind 绑定谷歌验证码
	Bind(adminID uint, secret, code string) error
	// Unbind 解绑谷歌验证码
	Unbind(adminID uint) error
}

type GoogleAuthenticatorServiceImpl struct {
}

func NewGoogleAuthenticatorServiceImpl() GoogleAuthenticatorService {
	return &GoogleAuthenticatorServiceImpl{}
}

// GenerateSecret 生成密钥
func (s *GoogleAuthenticatorServiceImpl) GenerateSecret(accountName string) (secret string, qrCodeURL string, err error) {
	// 获取应用名称（从配置中读取，如果没有则使用默认值）
	appName := facades.Config().GetString("app.name", "Goravel Admin")
	if appName == "" {
		appName = "Goravel Admin"
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      appName,
		AccountName: accountName,
	})
	if err != nil {
		return "", "", err
	}

	// 生成二维码URL
	qrCodeURL = key.URL()

	return key.Secret(), qrCodeURL, nil
}

// GenerateQRCodeImage 生成二维码图片（base64）
// 使用标准的TOTP格式（RFC 6238），完全兼容Google Authenticator
func (s *GoogleAuthenticatorServiceImpl) GenerateQRCodeImage(accountName, secret string) (string, error) {
	appName := facades.Config().GetString("app.name", "Goravel Admin")
	if appName == "" {
		appName = "Goravel Admin"
	}

	// 构建标准的otpauth URL（Google Authenticator标准格式）
	// 格式：otpauth://totp/Issuer:AccountName?secret=SECRET&issuer=Issuer
	otpURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", 
		appName, accountName, secret, appName)
	
	// 从URL创建key对象（这样可以确保格式完全符合标准）
	key, err := otp.NewKeyFromURL(otpURL)
	if err != nil {
		return "", fmt.Errorf("failed to create key from URL: %w", err)
	}

	// 生成二维码图片（200x200像素，Google Authenticator推荐尺寸）
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// 将图片编码为PNG
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode PNG: %w", err)
	}

	// 转换为base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/png;base64," + base64Str, nil
}

// Verify 验证验证码
func (s *GoogleAuthenticatorServiceImpl) Verify(secret, code string) bool {
	if secret == "" || code == "" {
		return false
	}
	return totp.Validate(code, secret)
}

// IsBound 检查管理员是否绑定了谷歌验证码
func (s *GoogleAuthenticatorServiceImpl) IsBound(adminID uint) (bool, error) {
	// 先检查列是否存在
	columns, err := facades.Schema().GetColumns("admins")
	if err != nil {
		return false, err
	}

	hasGoogleSecretColumn := false
	for _, column := range columns {
		if column.Name == "google_secret" {
			hasGoogleSecretColumn = true
			break
		}
	}

	// 如果列不存在，返回 false（未绑定）
	if !hasGoogleSecretColumn {
		return false, nil
	}

	// 列存在，检查是否有值
	count, err := facades.Orm().Query().Table("admins").
		Where("id", adminID).
		Where("google_secret IS NOT NULL").
		Where("google_secret != ?", "").
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetSecret 获取管理员的密钥（用于绑定确认）
func (s *GoogleAuthenticatorServiceImpl) GetSecret(adminID uint) (string, error) {
	var admin struct {
		GoogleSecret string
	}
	err := facades.Orm().Query().Table("admins").
		Select("google_secret").
		Where("id", adminID).
		First(&admin)
	if err != nil {
		return "", err
	}
	return admin.GoogleSecret, nil
}

// Bind 绑定谷歌验证码
func (s *GoogleAuthenticatorServiceImpl) Bind(adminID uint, secret, code string) error {
	// 先验证验证码是否正确
	if !s.Verify(secret, code) {
		return fmt.Errorf("invalid_code")
	}

	// 更新管理员的google_secret
	_, err := facades.Orm().Query().Table("admins").
		Where("id", adminID).
		Update("google_secret", secret)
	return err
}

// Unbind 解绑谷歌验证码
func (s *GoogleAuthenticatorServiceImpl) Unbind(adminID uint) error {
	_, err := facades.Orm().Query().Table("admins").
		Where("id", adminID).
		Update("google_secret", nil)
	return err
}

