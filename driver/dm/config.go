package dm

import (
	"fmt"

	"github.com/goravel/framework/contracts/config"

	"goravel/driver/dm/contracts"
)

type ConfigBuilder struct {
	config     config.Config
	connection string
}

func NewConfig(config config.Config, connection string) *ConfigBuilder {
	return &ConfigBuilder{
		config:     config,
		connection: connection,
	}
}

func (r *ConfigBuilder) Config() config.Config {
	return r.config
}

func (r *ConfigBuilder) Connection() string {
	return r.connection
}

func (r *ConfigBuilder) Readers() []contracts.FullConfig {
	configs := r.config.Get(fmt.Sprintf("database.connections.%s.read", r.connection))
	if readConfigs, ok := configs.([]contracts.Config); ok {
		return r.fillDefault(readConfigs)
	}

	return nil
}

func (r *ConfigBuilder) Writers() []contracts.FullConfig {
	configs := r.config.Get(fmt.Sprintf("database.connections.%s.write", r.connection))
	if writeConfigs, ok := configs.([]contracts.Config); ok {
		return r.fillDefault(writeConfigs)
	}

	return r.fillDefault([]contracts.Config{{}})
}

func (r *ConfigBuilder) fillDefault(configs []contracts.Config) []contracts.FullConfig {
	if len(configs) == 0 {
		return nil
	}

	var fullConfigs []contracts.FullConfig
	for _, cfg := range configs {
		fullConfig := contracts.FullConfig{
			Config:      cfg,
			Connection:  r.connection,
			Driver:      Name,
			NoLowerCase: r.config.GetBool(fmt.Sprintf("database.connections.%s.no_lower_case", r.connection)),
			Prefix:      r.config.GetString(fmt.Sprintf("database.connections.%s.prefix", r.connection)),
			Singular:    r.config.GetBool(fmt.Sprintf("database.connections.%s.singular", r.connection)),
		}
		if nameReplacer := r.config.Get(fmt.Sprintf("database.connections.%s.name_replacer", r.connection)); nameReplacer != nil {
			if replacer, ok := nameReplacer.(contracts.Replacer); ok {
				fullConfig.NameReplacer = replacer
			}
		}

		if fullConfig.Dsn == "" {
			fullConfig.Dsn = r.config.GetString(fmt.Sprintf("database.connections.%s.dsn", r.connection))
		}
		if fullConfig.Host == "" {
			fullConfig.Host = r.config.GetString(fmt.Sprintf("database.connections.%s.host", r.connection))
		}
		if fullConfig.Port == 0 {
			fullConfig.Port = r.config.GetInt(fmt.Sprintf("database.connections.%s.port", r.connection))
		}
		if fullConfig.Username == "" {
			fullConfig.Username = r.config.GetString(fmt.Sprintf("database.connections.%s.username", r.connection))
		}
		if fullConfig.Password == "" {
			fullConfig.Password = r.config.GetString(fmt.Sprintf("database.connections.%s.password", r.connection))
		}
		if fullConfig.Database == "" {
			fullConfig.Database = r.config.GetString(fmt.Sprintf("database.connections.%s.database", r.connection))
		}
		if fullConfig.Schema == "" {
			fullConfig.Schema = r.config.GetString(fmt.Sprintf("database.connections.%s.schema", r.connection))
		}
		fullConfigs = append(fullConfigs, fullConfig)
	}

	return fullConfigs
}
