//go:build !dm

package dm

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// Config keeps the same shape as the real DM dialector config.
type Config struct {
	DriverName        string
	DSN               string
	Conn              gorm.ConnPool
	DefaultStringSize uint
	GormMode          int
}

type unavailableDialector struct {
	reason error
}

func (u unavailableDialector) Name() string { return "dm" }

func (u unavailableDialector) Initialize(*gorm.DB) error { return u.reason }

func (u unavailableDialector) Migrator(*gorm.DB) gorm.Migrator { return nil }

func (u unavailableDialector) DataTypeOf(*schema.Field) string { return "" }

func (u unavailableDialector) DefaultValueOf(*schema.Field) clause.Expression { return nil }

func (u unavailableDialector) BindVarTo(clause.Writer, *gorm.Statement, interface{}) {}

func (u unavailableDialector) QuoteTo(clause.Writer, string) {}

func (u unavailableDialector) Explain(string, ...interface{}) string {
	return "dm dialector unavailable: build with -tags dm and install official dm driver"
}

func unavailableErr() error {
	return fmt.Errorf("dm dialector unavailable: build with -tags dm and install official dm Go driver package")
}

// Open returns a placeholder dialector when DM build tag is not enabled.
func Open(_ string) gorm.Dialector {
	return unavailableDialector{reason: unavailableErr()}
}

// New returns a placeholder dialector when DM build tag is not enabled.
func New(_ Config) gorm.Dialector {
	return unavailableDialector{reason: unavailableErr()}
}
