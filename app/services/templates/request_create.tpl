package admin

import (
	"goravel/app/http/trans"

	"github.com/goravel/framework/contracts/http"
)

type <<.RequestCreateName>> struct {
<<range .FormFields>>
	<<.FieldName>> <<.GoType>> `form:"<<.JsonName>>" json:"<<.JsonName>>"`
<<- end>>
}

func (r *<<.RequestCreateName>>) Authorize(ctx http.Context) error {
	return nil
}

func (r *<<.RequestCreateName>>) Rules(ctx http.Context) map[string]string {
	rules := map[string]string{
<<range .FormFields>>
		"<<.JsonName>>": "<<if .Required>>required<<end>><<if and .Required .Validators>>|<<end>><<range $i, $v := .Validators>><<if $i>>|<<end>><<$v>><<end>>",
<<- end>>
	}
	return rules
}

func (r *<<.RequestCreateName>>) Messages(ctx http.Context) map[string]string {
	return map[string]string{
<<range .FormFields>>
		"<<.JsonName>>.required": trans.Get(ctx, "validation_<<.Name>>_required"),
<<- end>>
	}
}

func (r *<<.RequestCreateName>>) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
<<range .FormFields>>
		"<<.JsonName>>": trans.Get(ctx, "validation_<<.Name>>"),
<<- end>>
	}
}