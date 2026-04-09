package dm

import (
	"fmt"
	"net/url"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database"
	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/contracts/process"
	"github.com/goravel/framework/contracts/testing/docker"
	"github.com/goravel/framework/errors"
	"gorm.io/gorm"

	"goravel/driver/dm/contracts"
)

const (
	Binding = "goravel.dm"
	Name    = "DM"
)

var _ contractsdriver.Driver = &DM{}

type DM struct {
	config  contracts.ConfigBuilder
	log     log.Log
	process process.Process
}

func NewDM(config config.Config, log log.Log, process process.Process, connection string) *DM {
	return &DM{
		config:  NewConfig(config, connection),
		log:     log,
		process: process,
	}
}

func (r *DM) Docker() (docker.DatabaseDriver, error) {
	if r.process == nil {
		return nil, errors.ProcessFacadeNotSet.SetModule(Name)
	}

	return nil, fmt.Errorf("dm docker driver is not implemented")
}

func (r *DM) Grammar() contractsdriver.Grammar {
	return NewGrammar(r.config.Writers()[0].Prefix)
}

func (r *DM) Pool() database.Pool {
	return database.Pool{
		Readers: r.fullConfigsToConfigs(r.config.Readers()),
		Writers: r.fullConfigsToConfigs(r.config.Writers()),
	}
}

func (r *DM) Processor() contractsdriver.Processor {
	return NewProcessor()
}

func (r *DM) fullConfigsToConfigs(fullConfigs []contracts.FullConfig) []database.Config {
	configs := make([]database.Config, len(fullConfigs))
	for i, fullConfig := range fullConfigs {
		configs[i] = database.Config{
			Connection:   fullConfig.Connection,
			Dsn:          fullConfig.Dsn,
			Database:     fullConfig.Database,
			Dialector:    fullConfigToDialector(fullConfig),
			Driver:       Name,
			Host:         fullConfig.Host,
			NameReplacer: fullConfig.NameReplacer,
			NoLowerCase:  fullConfig.NoLowerCase,
			Password:     fullConfig.Password,
			Port:         fullConfig.Port,
			Prefix:       fullConfig.Prefix,
			Schema:       fullConfig.Schema,
			Singular:     fullConfig.Singular,
			Username:     fullConfig.Username,
		}
	}

	return configs
}

func dsn(fullConfig contracts.FullConfig) string {
	if fullConfig.Dsn != "" {
		return fullConfig.Dsn
	}
	if fullConfig.Host == "" {
		return ""
	}

	escapedPassword := url.QueryEscape(fullConfig.Password)
	return fmt.Sprintf("%s:%s@%s:%d",
		fullConfig.Username, escapedPassword, fullConfig.Host, fullConfig.Port)
}

func fullConfigToDialector(fullConfig contracts.FullConfig) gorm.Dialector {
	dsn := dsn(fullConfig)
	if dsn == "" {
		return nil
	}

	// Matches the official DM dialector fields:
	// GormMode: dm-compatible=0; mysql-compatible=1
	return New(Config{
		DSN:      dsn,
		GormMode: fullConfig.GormMode,
	})
}
