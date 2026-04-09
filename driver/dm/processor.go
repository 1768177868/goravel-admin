package dm

import (
	"strings"

	"github.com/goravel/framework/contracts/database/driver"
	"github.com/spf13/cast"
)

var _ driver.Processor = &Processor{}

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (r Processor) ProcessColumns(dbColumns []driver.DBColumn) []driver.Column {
	var columns []driver.Column
	for _, dbColumn := range dbColumns {
		columns = append(columns, driver.Column{
			Autoincrement: strings.Contains(strings.ToUpper(dbColumn.Extra), "IDENTITY"),
			Collation:     dbColumn.Collation,
			Comment:       dbColumn.Comment,
			Default:       dbColumn.Default,
			Name:          dbColumn.Name,
			Nullable:      cast.ToBool(dbColumn.Nullable),
			Type:          dbColumn.Type,
			TypeName:      dbColumn.TypeName,
		})
	}

	return columns
}

func (r Processor) ProcessForeignKeys(dbForeignKeys []driver.DBForeignKey) []driver.ForeignKey {
	var foreignKeys []driver.ForeignKey

	for _, dbForeignKey := range dbForeignKeys {
		foreignKeys = append(foreignKeys, driver.ForeignKey{
			Name:           dbForeignKey.Name,
			Columns:        strings.Split(dbForeignKey.Columns, ","),
			ForeignSchema:  dbForeignKey.ForeignSchema,
			ForeignTable:   dbForeignKey.ForeignTable,
			ForeignColumns: strings.Split(dbForeignKey.ForeignColumns, ","),
			OnUpdate:       strings.ToLower(dbForeignKey.OnUpdate),
			OnDelete:       strings.ToLower(dbForeignKey.OnDelete),
		})
	}

	return foreignKeys
}

func (r Processor) ProcessIndexes(dbIndexes []driver.DBIndex) []driver.Index {
	var indexes []driver.Index
	for _, dbIndex := range dbIndexes {
		indexes = append(indexes, driver.Index{
			Columns: strings.Split(dbIndex.Columns, ","),
			Name:    strings.ToLower(dbIndex.Name),
			Type:    strings.ToLower(dbIndex.Type),
			Primary: dbIndex.Primary,
			Unique:  dbIndex.Unique,
		})
	}

	return indexes
}

func (r Processor) ProcessTypes(types []driver.Type) []driver.Type {
	return types
}
