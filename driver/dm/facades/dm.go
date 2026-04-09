package facades

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/driver"

	"goravel/driver/dm"
)

func Dm(connection string) (driver.Driver, error) {
	if dm.App == nil {
		return nil, fmt.Errorf("please register dm service provider")
	}

	instance, err := dm.App.MakeWith(dm.Binding, map[string]any{
		"connection": connection,
	})
	if err != nil {
		return nil, err
	}

	return instance.(*dm.DM), nil
}
