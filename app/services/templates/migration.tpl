package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type Create<<.ModelName>>Table struct {
}

func (m *Create<<.ModelName>>Table) Signature() string {
	return "create_<<.TableName>>_table"
}

func (m *Create<<.ModelName>>Table) Up() {
	facades.Schema().Create("<<.TableName>>", func(table schema.Blueprint) {
		table.ID()
<<range .Fields>>
		table.<<.MigrationMethod>>("<<.Name>>")<<if .Comment>>.Comment("<<.Comment>>")<<end>>
<<end>>
		table.Timestamps()
		table.SoftDeletes()
	})
}
func (m *Create<<.ModelName>>Table) Down() {
	facades.Schema().DropIfExists("<<.TableName>>")
}
