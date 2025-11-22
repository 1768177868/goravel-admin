package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20250101000001CreateDepartmentsTable struct {
}

func (r *M20250101000001CreateDepartmentsTable) Signature() string {
	return "20250101000001_create_departments_table"
}

func (r *M20250101000001CreateDepartmentsTable) Up() error {
	if !facades.Schema().HasTable("departments") {
		return facades.Schema().Create("departments", func(table schema.Blueprint) {
			table.BigIncrements("id")
			table.UnsignedBigInteger("parent_id").Default(0)
			table.String("name").Default("")
			table.String("code").Nullable()
			table.String("leader").Nullable()
			table.String("phone").Nullable()
			table.String("email").Nullable()
			table.UnsignedTinyInteger("status").Default(1)
			table.Integer("sort").Default(0)
			table.String("remark").Nullable()
			table.Timestamps()
			table.SoftDeletes()
			table.Comment("部门表")
		})
	}

	return nil
}

func (r *M20250101000001CreateDepartmentsTable) Down() error {
	return facades.Schema().DropIfExists("departments")
}
