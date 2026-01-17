package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type <<.RequestUpdateName>> struct {
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
	<<.FieldName>> *<<.GoType>> `form:"<<.JsonName>>" json:"<<.JsonName>>"`
<<- end>>
<<- end>>
}

func (r *<<.RequestUpdateName>>) Authorize(ctx http.Context) error {
	return nil
}

func (r *<<.RequestUpdateName>>) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
		"<<.JsonName>>": "<<if .Required>>required<<end>><<if and .Required .Validators>>|<<end>><<range $i, $v := .Validators>><<if $i>>|<<end>><<$v>><<end>>",
<<- end>>
<<- end>>
	}
	return rules
}

func (r *<<.RequestUpdateName>>) Messages(ctx http.Context) map[string]string {
	return map[string]string{
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
		"<<.JsonName>>.required": trans.Get(ctx, "validation_<<.Name>>_required"),
<<- end>>
<<- end>>
	}
}

func (r *<<.RequestUpdateName>>) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
<<range .FormFields>>
<<- if and (ne .Name "id") (ne .Name "created_at") (ne .Name "updated_at") (ne .Name "deleted_at")>>
		"<<.JsonName>>": trans.Get(ctx, "validation_<<.Name>>"),
<<- end>>
<<- end>>
	}
}