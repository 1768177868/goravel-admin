package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M<<.Timestamp>>Create<<.ModelName>>Table struct {
}

func (m *M<<.Timestamp>>Create<<.ModelName>>Table) Signature() string {
	return "<<.Timestamp>>_create_<<.TableName>>_table"
}

func (m *M<<.Timestamp>>Create<<.ModelName>>Table) Up() error {
	return facades.Schema().Create("<<.TableName>>", func(table schema.Blueprint) {
		table.ID()
<<range .Fields>>
		table.<<.MigrationMethod>>("<<.Name>>"<<if eq .DBType "decimal">>, <<.Precision>>, <<.Scale>><<end>>)<<if .Comment>>.Comment("<<.Comment>>")<<end>>
<<- end>>
		table.Timestamps()
		table.SoftDeletes()
	})
}
func (m *M<<.Timestamp>>Create<<.ModelName>>Table) Down() error {
	return facades.Schema().DropIfExists("<<.TableName>>")
}
