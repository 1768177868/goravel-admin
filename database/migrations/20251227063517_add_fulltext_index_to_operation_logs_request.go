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
	isDM := facades.Config().GetString("database.default") == "dm"

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
		if isDM {
			// DM: 优先尝试全文索引语法，失败则降级为普通索引（不阻断迁移）
			attempts := []string{
				"CREATE CONTEXT INDEX ft_request ON operation_logs(request)",
				"CREATE FULLTEXT INDEX ft_request ON operation_logs(request)",
				"CREATE INDEX ft_request ON operation_logs(request)",
			}
			for _, sql := range attempts {
				if _, err := query.Exec(sql); err == nil {
					return nil
				}
			}
			facades.Log().Warning("skip dm text index migration: all create-index attempts failed")
			return nil
		}

		// 检测数据库类型
		var dbType string
		if err := query.Raw("SELECT VERSION()").Scan(&dbType); err != nil {
			return err
		}

		if strings.Contains(dbType, "PostgreSQL") {
			// PostgreSQL: 先创建扩展（如果不存在）
			if _, err := query.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm"); err != nil {
				facades.Log().Warningf("skip pg_trgm index migration: cannot create extension, err=%v", err)
				return nil
			}
			// 创建 GIN 索引
			sql := "CREATE INDEX ft_request ON operation_logs USING GIN(request gin_trgm_ops)"
			if _, err := query.Exec(sql); err != nil {
				facades.Log().Warningf("skip pg_trgm index migration: cannot create index, err=%v", err)
			}
			return nil
		} else if strings.Contains(dbType, "MariaDB") {
			// MariaDB: 不支持 ngram 解析器，直接使用普通全文索引
			sql := "CREATE FULLTEXT INDEX ft_request ON operation_logs(request)"
			if _, err := query.Exec(sql); err != nil {
				facades.Log().Warningf("skip mariadb fulltext index migration: cannot create index, err=%v", err)
			}
			return nil
		} else {
			// MySQL: 尝试创建 ngram 全文索引，如果失败则使用普通全文索引
			sql := "CREATE FULLTEXT INDEX ft_request ON operation_logs(request) WITH PARSER ngram"
			_, err := query.Exec(sql)

			// 如果 ngram 解析器不可用，回退到普通全文索引
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "ngram") && (strings.Contains(errStr, "not defined") || strings.Contains(errStr, "not supported")) {
					// ngram 解析器不可用，使用普通全文索引
					sql = "CREATE FULLTEXT INDEX ft_request ON operation_logs(request)"
					if _, fallbackErr := query.Exec(sql); fallbackErr != nil {
						facades.Log().Warningf("skip mysql fulltext index migration: fallback create index failed, err=%v", fallbackErr)
					}
					return nil
				}
				facades.Log().Warningf("skip mysql fulltext index migration: create index failed, err=%v", err)
				return nil
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
	isDM := facades.Config().GetString("database.default") == "dm"

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
	if isDM {
		// DM 删除索引语法在不同版本可能差异，按顺序尝试（不阻断）
		dropAttempts := []string{
			"DROP INDEX ft_request",
			"DROP INDEX ft_request ON operation_logs",
		}
		for _, sql := range dropAttempts {
			if _, err := query.Exec(sql); err == nil {
				return nil
			}
		}
		facades.Log().Warning("skip drop dm index ft_request: all drop-index attempts failed")
		return nil
	}

	// 检测数据库类型
	var dbType string
	if err := query.Raw("SELECT VERSION()").Scan(&dbType); err != nil {
		return err
	}

	if strings.Contains(dbType, "PostgreSQL") {
		// PostgreSQL: 删除索引语法
		sql := "DROP INDEX IF EXISTS ft_request"
		if _, err := query.Exec(sql); err != nil {
			facades.Log().Warningf("skip drop pg index ft_request: err=%v", err)
		}
		return nil
	} else {
		// MySQL: 删除索引语法
		sql := "DROP INDEX ft_request ON operation_logs"
		if _, err := query.Exec(sql); err != nil {
			facades.Log().Warningf("skip drop mysql index ft_request: err=%v", err)
		}
		return nil
	}
}
