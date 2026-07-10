package utils

import (
	"context"
	"strings"

	appfacades "goravel/app/facades"

	"github.com/goravel/framework/contracts/database/orm"
)

// ApplyFulltextSearch 应用全文索引搜索条件
// column: 要搜索的字段名（如 "request", "content" 等）
// keyword: 搜索关键词
// query: ORM 查询对象
// 返回: 应用了搜索条件的查询对象
func ApplyFulltextSearch(query orm.Query, column, keyword string) orm.Query {
	if keyword == "" {
		return query
	}

	driver := strings.ToLower(query.Driver())

	isPostgreSQL := driver == "postgresql"
	isMySQLFamily := driver == "mysql" || driver == "mariadb"
	isDM := driver == "dm"

	// 判断是否应该使用全文索引
	// 对于短词（少于3个字符）或包含特殊字符的搜索，使用 LIKE/ILIKE
	// 对于长词，使用全文索引
	useFulltext := len(keyword) >= 3

	// 检查是否包含特殊字符（逗号、引号、括号等）
	specialChars := []string{",", "\"", "'", "(", ")", "[", "]", "{", "}", ":", ";", "=", "+", "-", "*", "/", "\\", "|", "&", "%", "$", "#", "@", "!", "?", "<", ">", "~", "`"}
	for _, char := range specialChars {
		if strings.Contains(keyword, char) {
			useFulltext = false
			break
		}
	}

	if useFulltext {
		if isPostgreSQL {
			// PostgreSQL: 使用 pg_trgm 相似度搜索（需要已创建 GIN 索引）
			// 使用 % 操作符进行相似度匹配，阈值默认 0.3
			return query.Where(column+" % ?", keyword)
		} else if isMySQLFamily {
			// MySQL: 使用 ngram 全文索引
			// 注意：需要确保字段已创建全文索引，索引名格式为 ft_{column}
			return query.Where("MATCH("+column+") AGAINST(? IN BOOLEAN MODE)", keyword)
		} else if isDM {
			// DM: 优先使用全文检索函数（可命中 CONTEXT/FULLTEXT 索引）
			// 说明：在未建全文索引或函数不可用时，调用方可退回 LIKE。
			return query.Where("CONTAINS("+column+", ?) > 0", keyword)
		} else {
			// other drivers: avoid MySQL-only AGAINST syntax.
			return query.Where(column+" LIKE ?", "%"+keyword+"%")
		}
	} else {
		// 短词或包含特殊字符：使用 LIKE/ILIKE
		if isPostgreSQL {
			// PostgreSQL: 使用 ILIKE（不区分大小写）
			return query.Where(column+" ILIKE ?", "%"+keyword+"%")
		} else {
			// MySQL / DM / others: 使用 LIKE
			return query.Where(column+" LIKE ?", "%"+keyword+"%")
		}
	}
}

// ShouldUseFulltextIndex 判断是否应该使用全文索引
// keyword: 搜索关键词
// 返回: true 表示应该使用全文索引，false 表示使用 LIKE/ILIKE
func ShouldUseFulltextIndex(keyword string) bool {
	if len(keyword) < 3 {
		return false
	}

	// 检查是否包含特殊字符
	specialChars := []string{",", "\"", "'", "(", ")", "[", "]", "{", "}", ":", ";", "=", "+", "-", "*", "/", "\\", "|", "&", "%", "$", "#", "@", "!", "?", "<", ">", "~", "`"}
	for _, char := range specialChars {
		if strings.Contains(keyword, char) {
			return false
		}
	}

	return true
}

// IsPostgreSQL 判断当前数据库是否为 PostgreSQL
func IsPostgreSQL(ctx context.Context) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	return strings.ToLower(appfacades.OrmQuery(ctx).Driver()) == "postgresql"
}
