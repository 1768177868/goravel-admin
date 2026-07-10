package facades

import (
	"context"

	"github.com/goravel/framework/contracts/database/orm"
)

func Orm() orm.Orm {
	return App().MakeOrm()
}

// OrmQuery returns an ORM query scoped to the given context (v1.18 WithContext).
func OrmQuery(ctx context.Context) orm.Query {
	return Orm().WithContext(ctx).Query()
}
