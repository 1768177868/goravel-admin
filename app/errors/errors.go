package errors

import (
	stderrors "errors"
	"fmt"
)

// 定义业务错误类型
var (
	// 认证相关错误
	ErrAccountDisabled      = NewBusinessError("account_disabled", "账号已被禁用")
	ErrPasswordError         = NewBusinessError("password_error", "密码错误")
	ErrNotLoggedIn          = NewBusinessError("not_logged_in", "未登录")
	ErrUsernameOrPasswordErr = NewBusinessError("username_or_password_error", "用户名或密码错误")
	ErrLoginFailed          = NewBusinessError("login_failed", "登录失败")

	// 验证相关错误
	ErrValidationFailed = NewBusinessError("validation_failed", "验证失败")
	ErrInvalidArgument  = NewBusinessError("invalid_argument", "无效的参数")

	// 资源相关错误
	ErrRecordNotFound = NewBusinessError("record_not_found", "记录不存在")
	ErrBlacklistNotFound = NewBusinessError("blacklist_not_found", "黑名单不存在")

	// IP 相关错误
	ErrIPAddressRequired    = NewBusinessError("ip_address_required", "IP地址不能为空")
	ErrInvalidCIDRFormat    = NewBusinessError("invalid_cidr_format", "CIDR格式错误")
	ErrInvalidIPRangeFormat = NewBusinessError("invalid_ip_range_format", "IP范围格式错误")
	ErrInvalidIPFormat      = NewBusinessError("invalid_ip_format", "IP格式错误")
	ErrInvalidIPRangeOrder  = NewBusinessError("invalid_ip_range_order", "IP范围顺序错误")
)

// BusinessError 业务错误类型
type BusinessError struct {
	Code    string
	Message string
	Err     error
}

// NewBusinessError 创建新的业务错误
func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回包装的错误
func (e *BusinessError) Unwrap() error {
	return e.Err
}

// WithError 包装底层错误
func (e *BusinessError) WithError(err error) *BusinessError {
	e.Err = err
	return e
}

// WithMessage 设置自定义消息
func (e *BusinessError) WithMessage(message string) *BusinessError {
	e.Message = message
	return e
}

// Is 检查错误是否匹配
func (e *BusinessError) Is(target error) bool {
	if t, ok := target.(*BusinessError); ok {
		return e.Code == t.Code
	}
	return stderrors.Is(e.Err, target)
}

// WrapError 包装错误并添加上下文信息
func WrapError(err error, code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsBusinessError 检查是否是业务错误
func IsBusinessError(err error) bool {
	_, ok := err.(*BusinessError)
	return ok
}

// GetBusinessError 获取业务错误
func GetBusinessError(err error) (*BusinessError, bool) {
	be, ok := err.(*BusinessError)
	return be, ok
}

