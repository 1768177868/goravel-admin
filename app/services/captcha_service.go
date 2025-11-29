package services

import (
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"

	"goravel/app/utils"
)

type CaptchaService interface {
	Enabled() bool
	Generate() (string, string, error)
	Verify(id, answer string) (bool, string)
}

type CaptchaServiceImpl struct {
	driver        base64Captcha.Driver
	store         base64Captcha.Store
	expireSeconds int
	initialized   bool
}

func NewCaptchaServiceImpl() CaptchaService {
	// 延迟初始化，不在构造函数中查询数据库
	// 这样可以避免在构建时访问数据库
	return &CaptchaServiceImpl{
		expireSeconds: 120, // 默认值
		initialized:   false,
	}
}

// initDriver 延迟初始化 driver 和 store
func (s *CaptchaServiceImpl) initDriver() {
	if s.initialized {
		return
	}

	// 从数据库读取验证码配置，如果不存在则使用默认值
	expireSeconds := utils.GetConfigValueInt("captcha", "captcha_expire", 120)
	if expireSeconds <= 0 {
		expireSeconds = 120
	}
	s.expireSeconds = expireSeconds

	s.driver = base64Captcha.NewDriverString(
		50,  // height
		180, // width
		5,   // noise count
		base64Captcha.OptionShowHollowLine|base64Captcha.OptionShowSineLine,
		5, // length
		"2345678ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz",
		nil,
		nil,
		nil,
	)

	s.store = base64Captcha.NewMemoryStore(1024, time.Duration(expireSeconds)*time.Second)
	s.initialized = true
}

func (s *CaptchaServiceImpl) Enabled() bool {
	// 从数据库读取验证码配置，如果不存在则使用默认值
	return utils.GetConfigValueBool("captcha", "captcha_enabled", false)
}

func (s *CaptchaServiceImpl) Generate() (string, string, error) {
	s.initDriver()
	c := base64Captcha.NewCaptcha(s.driver, s.store)
	id, b64s, _, err := c.Generate()
	if err != nil {
		return "", "", err
	}
	return id, b64s, nil
}

func (s *CaptchaServiceImpl) Verify(id, answer string) (bool, string) {
	if id == "" || strings.TrimSpace(answer) == "" {
		return false, "captcha_required"
	}

	s.initDriver()
	expected := s.store.Get(id, true)
	if expected == "" {
		return false, "captcha_expired"
	}

	if !strings.EqualFold(expected, strings.TrimSpace(answer)) {
		return false, "captcha_invalid"
	}

	return true, ""
}
