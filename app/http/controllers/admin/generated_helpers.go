package admin

import (
	stdhttp "net/http"
	"strings"

	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/response"
)

// ValidateGeneratedRequest validates a form request and returns a ready-to-send response on failure.
func ValidateGeneratedRequest(ctx http.Context, req http.FormRequest) http.Response {
	validationErrors, err := ctx.Request().ValidateRequest(req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if validationErrors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", validationErrors.All())
	}
	return nil
}

// HandleGeneratedServiceError maps BusinessError to HTTP responses and logs unexpected failures.
func HandleGeneratedServiceError(ctx http.Context, module string, status int, err error, attrs map[string]any) http.Response {
	if err == nil {
		return nil
	}
	if businessErr, ok := apperrors.GetBusinessError(err); ok {
		return response.Error(ctx, businessErrorStatus(businessErr.Code, status), businessErr.Code)
	}
	if status >= stdhttp.StatusInternalServerError {
		if attrs == nil {
			attrs = map[string]any{}
		}
		return response.ErrorWithLog(ctx, module, err, attrs)
	}
	return response.Error(ctx, status, err.Error())
}

func businessErrorStatus(code string, fallback int) int {
	switch {
	case code == "params_error" || code == "invalid_argument" || code == "validation_failed":
		return http.StatusBadRequest
	case strings.HasSuffix(code, "_required"):
		return http.StatusBadRequest
	case code == "record_not_found" || strings.HasSuffix(code, "_not_found"):
		return http.StatusNotFound
	case strings.HasSuffix(code, "_exists") || strings.HasSuffix(code, "_already_exists"):
		return http.StatusBadRequest
	case code == "password_encrypt_failed":
		return http.StatusInternalServerError
	case code == "too_many_requests":
		return http.StatusTooManyRequests
	case strings.HasPrefix(code, "role_protected_"):
		return http.StatusForbidden
	case strings.Contains(code, "_has_") || strings.HasPrefix(code, "protected_"):
		return http.StatusBadRequest
	case code == "account_disabled" || code == "forbidden":
		return http.StatusForbidden
	case code == "not_logged_in" || code == "username_or_password_error" || code == "unauthorized":
		return http.StatusUnauthorized
	case code == "query_failed" || code == "create_failed" || code == "update_failed" || code == "delete_failed" || code == "operation_failed":
		return http.StatusInternalServerError
	default:
		if fallback >= stdhttp.StatusInternalServerError {
			return http.StatusBadRequest
		}
		return fallback
	}
}
