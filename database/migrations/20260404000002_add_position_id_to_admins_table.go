package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M20260404000002AddPositionIdToAdminsTable struct {
}

func (r *M20260404000002AddPositionIdToAdminsTable) Signature() string {
	return "20260404000002_add_position_id_to_admins_table"
}

func (r *M20260404000002AddPositionIdToAdminsTable) Up() error {
	if !facades.Schema().HasTable("admins") {
		return nil
	}
	columns, err := facades.Schema().GetColumns("admins")
	if err != nil {
		return err
	}
	for _, column := range columns {
		if column.Name == "position_id" {
			return nil
		}
	}
	return facades.Schema().Table("admins", func(table schema.Blueprint) {
		table.UnsignedBigInteger("position_id").Nullable().Comment("岗位ID")
		table.Index("position_id")
	})
}

func (r *M20260404000002AddPositionIdToAdminsTable) Down() error {
	if !facades.Schema().HasTable("admins") {
		return nil
	}
	return facades.Schema().Table("admins", func(table schema.Blueprint) {
		table.DropColumn("position_id")
	})
}
