package dm

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

	// Official dm driver requires DSN to start with dm:// (see dm.DmConnector.parseDSN).
	var b strings.Builder
	b.WriteString("dm://")
	if fullConfig.Username != "" {
		b.WriteString(url.QueryEscape(fullConfig.Username))
		if fullConfig.Password != "" {
			b.WriteByte(':')
			b.WriteString(url.QueryEscape(fullConfig.Password))
		}
		b.WriteByte('@')
	}
	b.WriteString(fullConfig.Host)
	if fullConfig.Port > 0 {
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(fullConfig.Port))
	}
	// Optional path segment: DM *schema* (driver "catalog"). Instance/service names (e.g. DAMENG) are not valid here (-2103).
	// Only append when DB_SCHEMA is set; otherwise omit the path and use the login user's default schema.
	if fullConfig.Schema != "" {
		b.WriteByte('/')
		b.WriteString(fullConfig.Schema)
	}
	return b.String()
}

func fullConfigToDialector(fullConfig contracts.FullConfig) gorm.Dialector {
	dsn := dsn(fullConfig)
	if dsn == "" {
		return nil
	}

	// Matches the official DM dialector fields:
	// GormMode: dm-compatible=0; mysql-compatible=1
	return New(Config{
		DSN:             dsn,
		GormMode:        fullConfig.GormMode,
		SessionTimezone: fullConfig.SessionTimezone,
	})
}
