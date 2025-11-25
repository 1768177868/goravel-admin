package services

import (
	"strings"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/mojocn/base64Captcha"
)

type CaptchaService interface {
	Enabled() bool
	Generate() (string, string, error)
	Verify(id, answer string) (bool, string)
}

type CaptchaServiceImpl struct {
	driver base64Captcha.Driver
	store  base64Captcha.Store
}

func NewCaptchaServiceImpl() CaptchaService {
	expireSeconds := facades.Config().GetInt("admin.login_captcha_expire", 120)
	if expireSeconds <= 0 {
		expireSeconds = 120
	}

	driver := base64Captcha.NewDriverString(
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

	store := base64Captcha.NewMemoryStore(1024, time.Duration(expireSeconds)*time.Second)

	return &CaptchaServiceImpl{
		driver: driver,
		store:  store,
	}
}

func (s *CaptchaServiceImpl) Enabled() bool {
	return facades.Config().GetBool("admin.login_captcha_enabled", false)
}

func (s *CaptchaServiceImpl) Generate() (string, string, error) {
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

	expected := s.store.Get(id, true)
	if expected == "" {
		return false, "captcha_expired"
	}

	if !strings.EqualFold(expected, strings.TrimSpace(answer)) {
		return false, "captcha_invalid"
	}

	return true, ""
}
