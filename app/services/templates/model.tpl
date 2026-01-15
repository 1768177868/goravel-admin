package models

import (
	"github.com/goravel/framework/database/orm"
)

type <<.ModelName>> struct {
	orm.Model
	orm.SoftDeletes
<<range .Fields>>
	<<.FieldName>> <<.GoType>> `gorm:"<<.Name>>" json:"<<.JsonName>>"<<if .Comment>> comment:"<<.Comment>>"<<end>><<if .Relation>> json:"<<.Relation.Table>>_<<.Relation.DisplayField>>" gorm:"<<.Relation.Table>>;foreignKey:<<.Relation.ForeignKey>>"<<end>>`
<<- end>>
}

func (<<.ModelName>>) TableName() string {
	return "<<.TableName>>"
}

func (r *<<.ModelName>>) Serialize() map[string]any {
	return map[string]any{
		"id":         r.ID,
		"created_at": r.CreatedAt,
		"updated_at": r.UpdatedAt,
<<range .Fields>>
		"<<.JsonName>>": r.<<.FieldName>>,
<<end>>
	}
}

func (r *<<.ModelName>>) Deserialize(data map[string]any) {
<<range .Fields>>
	if val, ok := data["<<.JsonName>>"]; ok {
		r.<<.FieldName>> = val.(<<.GoType>>)
	}
<<end>>
}
