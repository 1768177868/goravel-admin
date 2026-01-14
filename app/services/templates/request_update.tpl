package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type <<.RequestUpdateName>> struct {
<<range .FormFields>>
	<<.FieldName>> *<<.GoType>> `form:"<<.JsonName>>" json:"<<.JsonName>>"`
<<end>>
}

func (r *<<.RequestUpdateName>>) Authorize(ctx http.Context) error {
	return nil
}

func (r *<<.RequestUpdateName>>) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{
<<range .FormFields>>
		"<<.JsonName>>": "<<if .Required>>required<<end>><<if and .Required .Validators>>|<<end>><<range $i, $v := .Validators>><<if $i>>|<<end>><<$v>><<end>>",
<<end>>
	}
	return rules
}

func (r *<<.RequestUpdateName>>) Messages(ctx http.Context) map[string]string {
	return map[string]string{
<<range .FormFields>>
		"<<.JsonName>>.required": trans.Get(ctx, "validation_<<.Name>>_required"),
<<end>>
	}
}

func (r *<<.RequestUpdateName>>) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
<<range .FormFields>>
		"<<.JsonName>>": trans.Get(ctx, "validation_<<.Name>>"),
<<end>>
	}
}