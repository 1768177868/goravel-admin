package models

import "github.com/goravel/framework/database/orm"

type SlowQueryLog struct {
	orm.Model
	TraceID       string  `gorm:"size:120;index;comment:链路ID" json:"trace_id"`
	SQLText       string  `gorm:"type:text;comment:原始SQL" json:"sql_text"`
	NormalizedSQL string  `gorm:"type:text;comment:归一化SQL" json:"normalized_sql"`
	SQLHash       string  `gorm:"size:64;index;comment:SQL摘要" json:"sql_hash"`
	DurationMS    float64 `gorm:"type:decimal(12,3);index;comment:耗时毫秒" json:"duration_ms"`
	RowsAffected  int64   `gorm:"comment:影响行数" json:"rows_affected"`
	Source        string  `gorm:"size:64;index;comment:来源" json:"source"`
	OccurredAt    string  `gorm:"size:32;index;comment:日志时间" json:"occurred_at"`
}
