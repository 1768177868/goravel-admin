package services

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/utils/errorlog"
)

// ShardingQueryConfig 分表查询配置
type ShardingQueryConfig struct {
	// BaseTableName 基础表名，如 "orders"
	BaseTableName string
	// GetColumns 获取表的所有列名（用于 UNION ALL 查询）
	// 返回格式：如 "id, order_no, user_id, created_at, updated_at, deleted_at"
	GetColumns func() string
	// BuildWhereClause 构建 WHERE 条件
	// 返回：WHERE 子句（不包含 WHERE 关键字）和参数列表
	// 例如：返回 ("user_id = ? AND status = ?", []any{1, "paid"})
	BuildWhereClause func(filters any) (string, []any)
	// GetAllowedOrderFields 获取允许排序的字段列表
	// 返回：字段名到 bool 的映射，如 map[string]bool{"id": true, "created_at": true}
	GetAllowedOrderFields func() map[string]bool
	// DefaultOrderBy 默认排序，格式：字段:方向，如 "created_at:desc"
	DefaultOrderBy string
	// ModuleName 模块名称，用于日志记录，如 "order"
	ModuleName string
}

// ShardingQueryService 分表查询服务接口
type ShardingQueryService interface {
	// QueryMultipleTables 查询多个分表（带分页）
	// tableNames: 分表名称列表
	// filters: 筛选条件（类型由具体实现决定）
	// page: 页码（从1开始）
	// pageSize: 每页数量
	// result: 查询结果（必须是指针类型）
	// 返回：结果列表、总数、错误
	QueryMultipleTables(tableNames []string, filters any, page, pageSize int, result any) (int64, error)

	// QueryMultipleTablesForExport 查询多个分表（不分页，用于导出）
	// tableNames: 分表名称列表
	// filters: 筛选条件（类型由具体实现决定）
	// result: 查询结果（必须是指针类型）
	// 返回：结果列表、错误
	QueryMultipleTablesForExport(tableNames []string, filters any, result any) error
}

type ShardingQueryServiceImpl struct {
	config ShardingQueryConfig
}

// NewShardingQueryService 创建分表查询服务
func NewShardingQueryService(config ShardingQueryConfig) ShardingQueryService {
	// 设置默认值
	if config.DefaultOrderBy == "" {
		config.DefaultOrderBy = "created_at:desc"
	}
	if config.ModuleName == "" {
		config.ModuleName = "sharding"
	}

	return &ShardingQueryServiceImpl{
		config: config,
	}
}

// QueryMultipleTables 查询多个分表（带分页）
func (s *ShardingQueryServiceImpl) QueryMultipleTables(tableNames []string, filters any, page, pageSize int, result any) (int64, error) {
	// 构建 WHERE 条件
	whereClause, whereConditions := s.config.BuildWhereClause(filters)
	if whereClause == "" {
		whereClause = "1=1"
	} else {
		whereClause = "1=1 AND " + whereClause
	}

	// 构建排序
	orderField, orderDir := s.parseOrderBy(filters)

	// 获取列名
	columnsStr := s.config.GetColumns()

	// 优化：每个分表先排序和限制，然后再合并（避免合并大量数据后再排序）
	// 计算每个分表需要查询的数量
	// 为了确保合并后有足够的数据进行分页，每个分表查询更多数据
	// 公式：limitPerTable = (page * pageSize) + pageSize，确保有足够数据
	offset := (page - 1) * pageSize
	limitPerTable := offset + pageSize + pageSize // 额外查询一页数据，确保有足够数据

	// 如果 limitPerTable 太大（超过10000），限制为10000，避免单个查询太慢
	if limitPerTable > 10000 {
		limitPerTable = 10000
	}

	// 构建 UNION ALL 查询
	// 过滤掉不存在的分表，避免查询错误
	var unionQueries []string
	var allArgs []any
	var existingTableNames []string

	for _, tableName := range tableNames {
		// 检查表是否存在，如果不存在则跳过
		if !facades.Schema().HasTable(tableName) {
			// 记录日志但不报错，因为某些月份的分表可能还没有创建
			errorlog.Record(context.Background(), s.config.ModuleName, "分表不存在，跳过查询", map[string]any{
				"table_name": tableName,
			}, "分表 %s 不存在，跳过查询", tableName)
			continue
		}

		// 表存在，添加到查询列表
		existingTableNames = append(existingTableNames, tableName)
		// 优化：每个分表先排序和限制，然后再合并
		// 使用子查询包装，确保每个分表先排序和限制
		query := fmt.Sprintf(
			"(SELECT %s FROM `%s` WHERE %s ORDER BY `%s` %s LIMIT %d)",
			columnsStr, tableName, whereClause, orderField, orderDir, limitPerTable,
		)
		unionQueries = append(unionQueries, query)
		// 每个查询都需要相同的参数
		allArgs = append(allArgs, whereConditions...)
	}

	if len(unionQueries) == 0 {
		return 0, nil
	}

	// 合并所有查询
	unionSQL := strings.Join(unionQueries, " UNION ALL ")

	// 优化：分别对每个分表执行 COUNT，然后相加（性能更好，可以利用索引）
	// 而不是对 UNION ALL 结果进行 COUNT（需要先合并所有数据）
	var total int64
	for _, tableName := range existingTableNames {
		// 每个分表使用相同的 WHERE 条件
		countSQL := fmt.Sprintf("SELECT COUNT(*) as total FROM `%s` WHERE %s", tableName, whereClause)
		var countResult struct {
			Total int64
		}
		// 使用对应的参数（每个分表使用相同的参数）
		args := whereConditions
		if err := facades.Orm().Query().Raw(countSQL, args...).Scan(&countResult); err != nil {
			errorlog.Record(context.Background(), s.config.ModuleName, "查询分表总数失败", map[string]any{
				"table_name": tableName,
				"error":      err.Error(),
			}, "查询分表 %s 总数失败: %v", tableName, err)
			// 如果某个分表查询失败，继续查询其他分表，但记录错误
			continue
		}
		total += countResult.Total
	}

	// 如果没有数据，直接返回
	if total == 0 {
		return 0, nil
	}

	// 分页查询（在外层再次排序和分页）
	// 注意：虽然每个分表已经排序，但合并后需要重新排序以确保全局顺序正确
	// 但由于每个分表已经限制了数量，合并后的数据量大大减少，排序会快很多
	paginatedSQL := fmt.Sprintf(
		"SELECT %s FROM (%s) as combined ORDER BY `%s` %s LIMIT ? OFFSET ?",
		columnsStr,
		unionSQL,
		orderField,
		orderDir,
	)

	// 添加 LIMIT 和 OFFSET 参数
	paginatedArgs := append(allArgs, pageSize, offset)

	// 执行查询
	if err := facades.Orm().Query().Raw(paginatedSQL, paginatedArgs...).Scan(result); err != nil {
		errorlog.Record(context.Background(), s.config.ModuleName, "查询列表失败", map[string]any{
			"table_count": len(tableNames),
			"page":        page,
			"page_size":   pageSize,
			"error":       err.Error(),
		}, "查询列表失败: %v", err)
		return 0, apperrors.ErrQueryFailed.WithError(err)
	}

	return total, nil
}

// QueryMultipleTablesForExport 查询多个分表（不分页，用于导出）
func (s *ShardingQueryServiceImpl) QueryMultipleTablesForExport(tableNames []string, filters any, result any) error {
	// 构建 WHERE 条件
	whereClause, whereConditions := s.config.BuildWhereClause(filters)
	if whereClause == "" {
		whereClause = "1=1"
	} else {
		whereClause = "1=1 AND " + whereClause
	}

	// 构建排序
	orderField, orderDir := s.parseOrderBy(filters)

	// 获取列名
	columnsStr := s.config.GetColumns()

	// 优化：每个分表先排序，然后再合并（对于导出，虽然需要所有数据，但先排序可以减少合并后的排序成本）
	// 构建 UNION ALL 查询
	// 过滤掉不存在的分表，避免查询错误
	var unionQueries []string
	var allArgs []any

	for _, tableName := range tableNames {
		// 检查表是否存在，如果不存在则跳过
		if !facades.Schema().HasTable(tableName) {
			// 记录日志但不报错，因为某些月份的分表可能还没有创建
			errorlog.Record(context.Background(), s.config.ModuleName, "分表不存在，跳过查询", map[string]any{
				"table_name": tableName,
			}, "分表 %s 不存在，跳过查询", tableName)
			continue
		}

		// 表存在，添加到查询列表
		// 优化：每个分表先排序，然后再合并
		// 使用子查询包装，确保每个分表先排序
		// 注意：导出需要所有数据，所以不限制数量，但先排序可以优化合并后的排序性能
		query := fmt.Sprintf(
			"(SELECT %s FROM `%s` WHERE %s ORDER BY `%s` %s)",
			columnsStr, tableName, whereClause, orderField, orderDir,
		)
		unionQueries = append(unionQueries, query)
		// 每个查询都需要相同的参数
		allArgs = append(allArgs, whereConditions...)
	}

	if len(unionQueries) == 0 {
		return nil
	}

	// 合并所有查询
	unionSQL := strings.Join(unionQueries, " UNION ALL ")

	// 导出查询（不分页，但需要排序）
	// 注意：虽然每个分表已经排序，但合并后需要重新排序以确保全局顺序正确
	// 但由于每个分表已经排序，合并后的排序会更快（归并排序）
	exportSQL := fmt.Sprintf(
		"SELECT %s FROM (%s) as combined ORDER BY `%s` %s",
		columnsStr,
		unionSQL,
		orderField,
		orderDir,
	)

	// 执行查询
	if err := facades.Orm().Query().Raw(exportSQL, allArgs...).Scan(result); err != nil {
		errorlog.Record(context.Background(), s.config.ModuleName, "导出查询失败", map[string]any{
			"table_count": len(tableNames),
			"error":       err.Error(),
		}, "导出查询失败: %v", err)
		return apperrors.ErrQueryFailed.WithError(err)
	}

	return nil
}

// parseOrderBy 解析排序字段
// 从 filters 中提取 OrderBy 字段（如果 filters 有 OrderBy 字段）
// 返回：排序字段名和方向
func (s *ShardingQueryServiceImpl) parseOrderBy(filters any) (string, string) {
	// 尝试从 filters 中提取 OrderBy 字段
	// 使用类型断言或反射来获取 OrderBy 字段
	// 这里使用一个辅助函数来处理
	orderBy := s.extractOrderBy(filters)
	if orderBy == "" {
		orderBy = s.config.DefaultOrderBy
	}

	orderParts := strings.Split(orderBy, ":")
	orderField := "created_at"
	orderDir := "desc"
	if len(orderParts) == 2 {
		field := orderParts[0]
		direction := strings.ToLower(orderParts[1])

		// 验证排序字段是否允许
		allowedFields := s.config.GetAllowedOrderFields()
		if allowedFields != nil && allowedFields[field] {
			orderField = field
			if direction == "asc" {
				orderDir = "asc"
			} else {
				orderDir = "desc"
			}
		}
	}

	return orderField, orderDir
}

// extractOrderBy 从 filters 中提取 OrderBy 字段
// 支持结构体和 map[string]any
func (s *ShardingQueryServiceImpl) extractOrderBy(filters any) string {
	if filters == nil {
		return ""
	}

	// 尝试类型断言为 map[string]any
	if m, ok := filters.(map[string]any); ok {
		if orderBy, ok := m["OrderBy"].(string); ok {
			return orderBy
		}
	}

	// 使用反射从结构体中提取 OrderBy 字段
	v := reflect.ValueOf(filters)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		field := v.FieldByName("OrderBy")
		if field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
	}

	return ""
}
