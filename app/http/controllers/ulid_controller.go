package controllers

import (
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/http/response"
	"goravel/app/utils"
)

type UlidController struct {
}

func NewUlidController() *UlidController {
	return &UlidController{}
}

// Generate 生成 ULID
func (r *UlidController) Generate(ctx http.Context) http.Response {
	ulid := utils.GenerateULID()
	return response.Success(ctx, "generate_success", http.Json{
		"ulid": ulid,
	})
}

// Parse 解析 ULID（获取时间信息）
func (r *UlidController) Parse(ctx http.Context) http.Response {
	ulidStr := ctx.Request().Query("ulid", "")
	if ulidStr == "" {
		return response.Error(ctx, http.StatusBadRequest, "ulid_required")
	}

	// 验证 ULID 是否有效
	if !utils.IsValidULID(ulidStr) {
		return response.Error(ctx, http.StatusBadRequest, "invalid_ulid")
	}

	// 解析时间
	t, err := utils.ParseULIDTime(ulidStr)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "parse_failed")
	}

	// 获取时间戳
	timestamp, err := utils.GetULIDTimestamp(ulidStr)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, "parse_failed")
	}

	// 格式化时间字符串
	timeString, _ := utils.ParseULIDTimeString(ulidStr, "2006-01-02 15:04:05")

	return response.Success(ctx, "parse_success", http.Json{
		"ulid":        ulidStr,
		"time":        t.Format("2006-01-02 15:04:05"),
		"timestamp":   timestamp,
		"time_iso":    t.Format(time.RFC3339),
		"time_string": timeString,
	})
}
