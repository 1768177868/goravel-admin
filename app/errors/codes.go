package errors

// 错误码定义
// 格式: XXYYYY
// XX: 模块 (10-认证, 20-权限, 30-验证, 40-业务, 50-系统)
// YYYY: 具体错误

const (
	// ==================== 认证模块 (10xxx) ====================

	// ErrCodeInvalidCredentials 用户名或密码错误
	ErrCodeInvalidCredentials = 10001
	// ErrCodeAccountDisabled 账号已被禁用
	ErrCodeAccountDisabled = 10002
	// ErrCodeTokenExpired Token 已过期
	ErrCodeTokenExpired = 10003
	// ErrCodeTokenRevoked Token 已被撤销
	ErrCodeTokenRevoked = 10004
	// ErrCodeCaptchaError 验证码错误
	ErrCodeCaptchaError = 10005
	// ErrCodeLoginLimitExceeded 登录尝试次数超限
	ErrCodeLoginLimitExceeded = 10006
	// ErrCodeTokenInvalid Token 无效
	ErrCodeTokenInvalid = 10007

	// ==================== 权限模块 (20xxx) ====================

	// ErrCodeNoPermission 无权限访问
	ErrCodeNoPermission = 20001
	// ErrCodeResourceNotFound 资源不存在
	ErrCodeResourceNotFound = 20002
	// ErrCodeAccessDenied 访问被拒绝
	ErrCodeAccessDenied = 20003

	// ==================== 验证模块 (30xxx) ====================

	// ErrCodeValidationFailed 参数验证失败
	ErrCodeValidationFailed = 30001
	// ErrCodeDataExists 数据已存在
	ErrCodeDataExists = 30002
	// ErrCodeDataNotFound 数据不存在
	ErrCodeDataNotFound = 30003
	// ErrCodeInvalidFormat 格式无效
	ErrCodeInvalidFormat = 30004

	// ==================== 业务模块 (40xxx) ====================

	// ErrCodeOperationFailed 操作失败
	ErrCodeOperationFailed = 40001
	// ErrCodeDeleteFailed 删除失败
	ErrCodeDeleteFailed = 40002
	// ErrCodeUpdateFailed 更新失败
	ErrCodeUpdateFailed = 40003
	// ErrCodeCreateFailed 创建失败
	ErrCodeCreateFailed = 40004
	// ErrCodeUploadFailed 上传失败
	ErrCodeUploadFailed = 40005
	// ErrCodeExportFailed 导出失败
	ErrCodeExportFailed = 40006

	// ==================== 系统模块 (50xxx) ====================

	// ErrCodeInternalError 服务器内部错误
	ErrCodeInternalError = 50001
	// ErrCodeDatabaseError 数据库错误
	ErrCodeDatabaseError = 50002
	// ErrCodeCacheError 缓存错误
	ErrCodeCacheError = 50003
	// ErrCodeQueueError 队列错误
	ErrCodeQueueError = 50004
	// ErrCodeThirdPartyError 第三方服务错误
	ErrCodeThirdPartyError = 50005
)

// ErrorMessages 错误码对应的消息 (用于不支持 i18n 的场景)
var ErrorMessages = map[int]string{
	// 认证
	ErrCodeInvalidCredentials: "用户名或密码错误",
	ErrCodeAccountDisabled:    "账号已被禁用",
	ErrCodeTokenExpired:       "Token 已过期",
	ErrCodeTokenRevoked:       "Token 已被撤销",
	ErrCodeCaptchaError:       "验证码错误",
	ErrCodeLoginLimitExceeded: "登录尝试次数超限",
	ErrCodeTokenInvalid:       "Token 无效",

	// 权限
	ErrCodeNoPermission:     "无权限访问",
	ErrCodeResourceNotFound: "资源不存在",
	ErrCodeAccessDenied:     "访问被拒绝",

	// 验证
	ErrCodeValidationFailed: "参数验证失败",
	ErrCodeDataExists:       "数据已存在",
	ErrCodeDataNotFound:     "数据不存在",
	ErrCodeInvalidFormat:    "格式无效",

	// 业务
	ErrCodeOperationFailed: "操作失败",
	ErrCodeDeleteFailed:    "删除失败",
	ErrCodeUpdateFailed:    "更新失败",
	ErrCodeCreateFailed:    "创建失败",
	ErrCodeUploadFailed:    "上传失败",
	ErrCodeExportFailed:    "导出失败",

	// 系统
	ErrCodeInternalError:   "服务器内部错误",
	ErrCodeDatabaseError:   "数据库错误",
	ErrCodeCacheError:      "缓存错误",
	ErrCodeQueueError:      "队列错误",
	ErrCodeThirdPartyError: "第三方服务错误",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code int) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}

// IsAuthError 是否为认证错误
func IsAuthError(code int) bool {
	return code >= 10000 && code < 20000
}

// IsPermissionError 是否为权限错误
func IsPermissionError(code int) bool {
	return code >= 20000 && code < 30000
}

// IsValidationError 是否为验证错误
func IsValidationError(code int) bool {
	return code >= 30000 && code < 40000
}

// IsBusinessError 是否为业务错误
func IsBusinessErrorCode(code int) bool {
	return code >= 40000 && code < 50000
}

// IsSystemError 是否为系统错误
func IsSystemError(code int) bool {
	return code >= 50000 && code < 60000
}
