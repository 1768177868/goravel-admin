package helpers

// PaginateSlice 对切片进行分页处理
// 返回分页后的切片和总数
func PaginateSlice[T any](slice []T, page, pageSize int) ([]T, int64) {
	total := int64(len(slice))
	if total == 0 {
		return []T{}, 0
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	// 如果起始位置超出范围，返回空切片
	if start >= len(slice) {
		return []T{}, total
	}

	// 如果结束位置超出范围，截取到末尾
	if end > len(slice) {
		end = len(slice)
	}

	return slice[start:end], total
}

// ValidatePagination 验证并规范化分页参数
// 返回规范化后的 page 和 pageSize
func ValidatePagination(page, pageSize int) (int, int) {
	// 默认值
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// 最大限制
	const maxPageSize = 100
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	return page, pageSize
}
