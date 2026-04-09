package apidoc

// Swagger 命名约定（推荐）：
// 1) XxxListResponse  = Success + XxxListData
// 2) XxxDetailResponse = Success + XxxDetailData
// 3) XxxListData 至少包含 List + Pagination（可按模块扩展）

// Success 成功响应公共外层（与 response.Success 的 JSON 一致），各模块用嵌入 + 自定义 Data 组合 Swagger 类型。
type Success struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"`
}

// Pagination 分页元信息（用于 data 内层复用）。
type Pagination struct {
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// Error 错误响应公共结构（与 response.Error 的 JSON 一致）。
type Error struct {
	Code      int    `json:"code" example:"404"`
	ErrorCode string `json:"error_code" example:"record_not_found"`
	Message   string `json:"message" example:"记录不存在"`
	TraceID   string `json:"trace_id,omitempty" example:"01knrfw7dsrxc3dxbbzk6bkv3j"`
}
