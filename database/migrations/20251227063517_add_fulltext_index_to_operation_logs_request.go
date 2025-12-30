package migrations

import (
	"strings"

	"github.com/goravel/framework/facades"
)

type M20251227063517AddFulltextIndexToOperationLogsRequest struct{}

// Signature The unique signature for the migration.
func (r *M20251227063517AddFulltextIndexToOperationLogsRequest) Signature() string {
	return "20251227063517_add_fulltext_index_to_operation_logs_request"
}

// Up Run the migrations.
func (r *M20251227063517AddFulltextIndexToOperationLogsRequest) Up() error {
	if !facades.Schema().HasTable("operation_logs") {
		return nil
	}

	// 检查索引是否已存在
	indexes, err := facades.Schema().GetIndexes("operation_logs")
	if err != nil {
		return err
	}

	hasIndex := false
	for _, index := range indexes {
		if index.Name == "ft_request" {
			hasIndex = true
			break
		}
	}

	if !hasIndex {
		query := facades.Orm().Query()

		// 检测数据库类型
		var dbType string
		if err := query.Raw("SELECT VERSION()").Scan(&dbType); err != nil {
			return err
		}

		if strings.Contains(dbType, "PostgreSQL") {
			// PostgreSQL: 先创建扩展（如果不存在）
			var dummy string
			_ = query.Raw("CREATE EXTENSION IF NOT EXISTS pg_trgm").Scan(&dummy)
			// 创建 GIN 索引
			sql := "CREATE INDEX ft_request ON operation_logs USING GIN(request gin_trgm_ops)"
			return query.Raw(sql).Scan(&dummy)
		} else if strings.Contains(dbType, "MariaDB") {
			// MariaDB: 不支持 ngram 解析器，直接使用普通全文索引
			var dummy string
			sql := "CREATE FULLTEXT INDEX ft_request ON operation_logs(request)"
			return query.Raw(sql).Scan(&dummy)
		} else {
			// MySQL: 尝试创建 ngram 全文索引，如果失败则使用普通全文索引
			var dummy string
			sql := "CREATE FULLTEXT INDEX ft_request ON operation_logs(request) WITH PARSER ngram"
			err := query.Raw(sql).Scan(&dummy)

			// 如果 ngram 解析器不可用，回退到普通全文索引
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "ngram") && (strings.Contains(errStr, "not defined") || strings.Contains(errStr, "not supported")) {
					// ngram 解析器不可用，使用普通全文索引
					sql = "CREATE FULLTEXT INDEX ft_request ON operation_logs(request)"
					return query.Raw(sql).Scan(&dummy)
				}
				return err
			}

			return nil
		}
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20251227063517AddFulltextIndexToOperationLogsRequest) Down() error {
	if !facades.Schema().HasTable("operation_logs") {
		return nil
	}

	// 检查索引是否存在
	indexes, err := facades.Schema().GetIndexes("operation_logs")
	if err != nil {
		return err
	}

	hasIndex := false
	for _, index := range indexes {
		if index.Name == "ft_request" {
			hasIndex = true
			break
		}
	}

	if !hasIndex {
		return nil
	}

	query := facades.Orm().Query()

	// 检测数据库类型
	var dbType string
	if err := query.Raw("SELECT VERSION()").Scan(&dbType); err != nil {
		return err
	}

	var dummy string
	if strings.Contains(dbType, "PostgreSQL") {
		// PostgreSQL: 删除索引语法
		sql := "DROP INDEX IF EXISTS ft_request"
		return query.Raw(sql).Scan(&dummy)
	} else {
		// MySQL: 删除索引语法
		sql := "DROP INDEX ft_request ON operation_logs"
		return query.Raw(sql).Scan(&dummy)
	}
}
