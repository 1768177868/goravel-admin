package api

import (
	"fmt"
	nethttp "net/http"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
	"goravel/app/http/trans"
	"goravel/app/utils"
)

// timeRangeErrorResponse 将时间范围错误转换为统一响应
func timeRangeErrorResponse(ctx http.Context, err error) http.Response {
	if err == nil {
		return nil
	}

	if timeRangeErr, ok := err.(*utils.TimeRangeError); ok {
		message := trans.Get(ctx, timeRangeErr.Key)
		if timeRangeErr.Params != nil {
			for key, value := range timeRangeErr.Params {
				placeholder := fmt.Sprintf("{%s}", key)
				message = strings.ReplaceAll(message, placeholder, fmt.Sprintf("%v", value))
			}
		}
		return response.Error(ctx, nethttp.StatusBadRequest, message)
	}

	return nil
}
