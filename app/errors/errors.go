package errors

import (
	stderrors "errors"
	"fmt"
)

// 定义业务错误类型
var (
	// 认证相关错误
	ErrAccountDisabled       = NewBusinessError("account_disabled", "账号已被禁用")
	ErrPasswordError         = NewBusinessError("password_error", "密码错误")
	ErrNotLoggedIn           = NewBusinessError("not_logged_in", "未登录")
	ErrUsernameOrPasswordErr = NewBusinessError("username_or_password_error", "用户名或密码错误")
	ErrLoginFailed           = NewBusinessError("login_failed", "登录失败")

	// 验证相关错误
	ErrValidationFailed = NewBusinessError("validation_failed", "验证失败")
	ErrInvalidArgument  = NewBusinessError("invalid_argument", "无效的参数")

	// 资源相关错误
	ErrRecordNotFound       = NewBusinessError("record_not_found", "记录不存在")
	ErrBlacklistNotFound    = NewBusinessError("blacklist_not_found", "黑名单不存在")
	ErrNotificationNotFound = NewBusinessError("notification_not_found", "通知不存在")
	ErrAdminNotFound        = NewBusinessError("admin_not_found", "管理员不存在")
	ErrRoleNotFound         = NewBusinessError("role_not_found", "角色不存在")
	ErrMenuNotFound         = NewBusinessError("menu_not_found", "菜单不存在")
	ErrPermissionNotFound   = NewBusinessError("permission_not_found", "权限不存在")
	ErrDepartmentNotFound   = NewBusinessError("department_not_found", "部门不存在")
	ErrDictionaryNotFound   = NewBusinessError("dictionary_not_found", "字典不存在")
	ErrLogNotFound          = NewBusinessError("log_not_found", "日志不存在")

	// IP 相关错误
	ErrIPAddressRequired    = NewBusinessError("ip_address_required", "IP地址不能为空")
	ErrInvalidCIDRFormat    = NewBusinessError("invalid_cidr_format", "CIDR格式错误")
	ErrInvalidIPRangeFormat = NewBusinessError("invalid_ip_range_format", "IP范围格式错误")
	ErrInvalidIPFormat      = NewBusinessError("invalid_ip_format", "IP格式错误")
	ErrInvalidIPRangeOrder  = NewBusinessError("invalid_ip_range_order", "IP范围顺序错误")

	// 参数相关错误
	ErrIDRequired       = NewBusinessError("id_required", "ID不能为空")
	ErrIDsRequired      = NewBusinessError("ids_required", "IDs不能为空")
	ErrParamsError      = NewBusinessError("params_error", "参数错误")
	ErrParamsRequired   = NewBusinessError("params_required", "参数不能为空")
	ErrFileRequired     = NewBusinessError("file_required", "文件不能为空")
	ErrFilePathRequired = NewBusinessError("file_path_required", "文件路径不能为空")
	ErrCodeRequired     = NewBusinessError("code_required", "验证码不能为空")
	ErrTokenIDRequired  = NewBusinessError("token_id_required", "Token ID不能为空")
	ErrTokenIDsRequired = NewBusinessError("token_ids_required", "Token IDs不能为空")
	ErrUserIDRequired   = NewBusinessError("user_id_required", "用户ID不能为空")
	ErrChunkIDRequired  = NewBusinessError("chunk_id_required", "分片ID不能为空")
	ErrFilenameRequired = NewBusinessError("filename_required", "文件名不能为空")

	// 附件相关错误
	ErrChunkUploadOnlyLocalStorage = NewBusinessError("chunk_upload_only_local_storage", "大文件分片上传仅支持本地存储")
	ErrInvalidChunkIndex           = NewBusinessError("invalid_chunk_index", "分片索引无效")
	ErrInvalidTotalChunks          = NewBusinessError("invalid_total_chunks", "总分片数无效")
	ErrInvalidTotalSize            = NewBusinessError("invalid_total_size", "总大小无效")
	ErrInvalidChunkSize            = NewBusinessError("invalid_chunk_size", "分片大小无效")
	ErrChunkFileRequired           = NewBusinessError("chunk_file_required", "分片文件不能为空")
	ErrInvalidAction               = NewBusinessError("invalid_action", "无效的操作")

	// 数据存在性错误
	ErrUsernameExists                = NewBusinessError("username_exists", "用户名已存在")
	ErrMenuSlugExists                = NewBusinessError("menu_slug_exists", "菜单标识已存在")
	ErrRoleNameExists                = NewBusinessError("role_name_exists", "角色名称已存在")
	ErrRoleSlugExists                = NewBusinessError("role_slug_exists", "角色标识已存在")
	ErrPermissionNameExists          = NewBusinessError("permission_name_exists", "权限名称已存在")
	ErrPermissionSlugExists          = NewBusinessError("permission_slug_exists", "权限标识已存在")
	ErrPermissionNameOrSlugExists    = NewBusinessError("permission_name_or_slug_exists", "权限名称或标识已存在")
	ErrPermissionNameAndSlugRequired = NewBusinessError("permission_name_and_slug_required", "权限名称和标识不能为空")

	// 业务逻辑错误
	ErrMenuHasChildren                 = NewBusinessError("menu_has_children", "菜单存在子菜单，无法删除")
	ErrDepartmentHasChildren           = NewBusinessError("department_has_children", "部门存在子部门，无法删除")
	ErrDepartmentHasAdmins             = NewBusinessError("department_has_admins", "部门存在管理员，无法删除")
	ErrGoogleAuthenticatorNotBound     = NewBusinessError("google_authenticator_not_bound", "未绑定谷歌验证器")
	ErrGoogleAuthenticatorAlreadyBound = NewBusinessError("google_authenticator_already_bound", "已绑定谷歌验证器")
	ErrGoogleCodeInvalid               = NewBusinessError("google_code_invalid", "谷歌验证码无效")
	ErrGoogleCodeRequired              = NewBusinessError("google_code_required", "谷歌验证码不能为空")
	ErrSecretAndCodeRequired           = NewBusinessError("secret_and_code_required", "密钥和验证码不能为空")
	ErrOldPasswordError                = NewBusinessError("old_password_error", "旧密码错误")
	ErrInvalidTokenID                  = NewBusinessError("invalid_token_id", "无效的Token ID")
	ErrInvalidTokenIDs                 = NewBusinessError("invalid_token_ids", "无效的Token IDs")
	ErrInvalidUserID                   = NewBusinessError("invalid_user_id", "无效的用户ID")
	ErrDictionaryTypeRequired          = NewBusinessError("dictionary_type_required", "字典类型不能为空")
	ErrConfigGroupRequired             = NewBusinessError("config_group_required", "配置组不能为空")
	ErrConfigsRequired                 = NewBusinessError("configs_required", "配置不能为空")
	ErrEmailConfigRequired             = NewBusinessError("email_config_required", "邮箱配置不能为空")
	ErrOptionTypeRequired              = NewBusinessError("option_type_required", "选项类型不能为空")
	ErrInvalidOptionType               = NewBusinessError("invalid_option_type", "无效的选项类型")

	// 操作相关错误
	ErrCreateFailed          = NewBusinessError("create_failed", "创建失败")
	ErrUpdateFailed          = NewBusinessError("update_failed", "更新失败")
	ErrQueryFailed           = NewBusinessError("query_failed", "查询失败")
	ErrDeleteFailed          = NewBusinessError("delete_failed", "删除失败")
	ErrPasswordEncryptFailed = NewBusinessError("password_encrypt_failed", "密码加密失败")

	// 资源不存在错误（其他资源错误已在资源相关错误部分定义）
	ErrTokenNotFound      = NewBusinessError("token_not_found", "Token不存在")
	ErrUnauthorized       = NewBusinessError("unauthorized", "未授权")
	ErrTokenRefreshFailed = NewBusinessError("token_refresh_failed", "Token刷新失败")
	ErrUserNotFound       = NewBusinessError("user_not_found", "用户不存在")

	// 权限保护错误
	ErrAdminProtectedCannotDisable   = NewBusinessError("admin_protected_cannot_disable", "受保护的管理员不能禁用")
	ErrAdminCannotModifyRoles        = NewBusinessError("admin_cannot_modify_roles", "不能修改管理员角色")
	ErrAdminProtectedCannotDelete    = NewBusinessError("admin_protected_cannot_delete", "受保护的管理员不能删除")
	ErrAdminCannotDeleteSelf         = NewBusinessError("admin_cannot_delete_self", "不能删除自己")
	ErrRoleProtectedCannotModifySlug = NewBusinessError("role_protected_cannot_modify_slug", "受保护的角色不能修改标识")
	ErrRoleProtectedCannotDisable    = NewBusinessError("role_protected_cannot_disable", "受保护的角色不能禁用")
	ErrRoleProtectedCannotDelete     = NewBusinessError("role_protected_cannot_delete", "受保护的角色不能删除")

	// 测试相关错误
	ErrTraceTestWarning = NewBusinessError("trace_test_warning", "追踪测试警告")
	ErrTraceTestError   = NewBusinessError("trace_test_error", "追踪测试错误")
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
