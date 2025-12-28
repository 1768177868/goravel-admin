package utils

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	"goravel/app/utils/errorlog"
)

var (
	gormDBInstance *gorm.DB
	gormDBOnce     sync.Once
	gormDBErr      error
)

// GetGormDB 获取原生 GORM DB 实例
// 注意：Goravel 框架的 ORM 不直接暴露底层 GORM DB 实例（DB() 方法返回的是 *sql.DB）
// 因此需要根据配置创建新的 GORM 连接，用于 GORM Sharding 等需要原生 GORM 功能的场景
// 使用单例模式，确保只创建一次连接
func GetGormDB() (*gorm.DB, error) {
	gormDBOnce.Do(func() {
		// 尝试从框架获取连接（虽然框架不直接暴露，但保留尝试逻辑以便未来框架更新时自动适配）
		if db, err := tryGetGormFromFramework(); err == nil && db != nil {
			facades.Log().Infof("成功从框架获取 GORM DB 实例")
			gormDBInstance = db
			gormDBErr = nil
			return
		}
		// 框架无法直接提供 GORM DB，创建新连接
		facades.Log().Infof("框架未直接暴露 GORM DB 实例，将根据配置创建新连接")
		gormDBInstance, gormDBErr = createGormConnection()
	})
	return gormDBInstance, gormDBErr
}

// tryGetGormFromFramework 尝试从框架的 ORM 获取底层 GORM DB 实例
// 目前框架的 DB() 方法返回 *sql.DB 而非 *gorm.DB，因此此方法通常无法获取到 GORM DB
// 保留此方法以便未来框架更新时自动适配
func tryGetGormFromFramework() (*gorm.DB, error) {
	orm := facades.Orm()
	if orm == nil {
		return nil, fmt.Errorf("框架 ORM 不可用")
	}

	ormValue := reflect.ValueOf(orm)

	// 尝试通过反射查找可能的 GORM DB 字段
	if ormValue.Kind() == reflect.Ptr && !ormValue.IsNil() {
		elemValue := ormValue.Elem()
		if elemValue.Kind() == reflect.Struct {
			for i := 0; i < elemValue.NumField(); i++ {
				field := elemValue.Field(i)
				fieldType := elemValue.Type().Field(i)

				// 检查字段类型是否是 *gorm.DB
				if fieldType.Type.String() == "*gorm.DB" || fieldType.Type.String() == "gorm.DB" {
					facades.Log().Infof("找到可能的 GORM DB 字段: %s", fieldType.Name)
					if field.CanInterface() {
						if db, ok := field.Interface().(*gorm.DB); ok && db != nil {
							facades.Log().Infof("成功从字段 %s 获取 GORM DB 实例", fieldType.Name)
							return db, nil
						}
					}
				}

				// 尝试递归查找（如果字段是结构体或指针）
				if field.Kind() == reflect.Ptr && !field.IsNil() {
					fieldElem := field.Elem()
					if fieldElem.Kind() == reflect.Struct {
						for j := 0; j < fieldElem.NumField(); j++ {
							subField := fieldElem.Field(j)
							subFieldType := fieldElem.Type().Field(j)
							if subFieldType.Type.String() == "*gorm.DB" {
								facades.Log().Infof("在嵌套字段 %s.%s 中找到 GORM DB", fieldType.Name, subFieldType.Name)
								if subField.CanInterface() {
									if db, ok := subField.Interface().(*gorm.DB); ok && db != nil {
										facades.Log().Infof("成功从嵌套字段获取 GORM DB 实例")
										return db, nil
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 尝试通过 Query() 方法获取查询对象，然后从查询对象获取 GORM DB
	if queryMethod := ormValue.MethodByName("Query"); queryMethod.IsValid() {
		queryResults := queryMethod.Call(nil)
		if len(queryResults) > 0 {
			queryValue := queryResults[0]
			if queryValue.Kind() == reflect.Interface && !queryValue.IsNil() {
				if queryElem := queryValue.Elem(); queryElem.IsValid() {
					queryElemType := queryElem.Type()
					// 尝试查找 GORM DB 字段
					if queryElemType.Kind() == reflect.Struct {
						for i := 0; i < queryElemType.NumField(); i++ {
							field := queryElemType.Field(i)
							if field.Type.String() == "*gorm.DB" {
								fieldValue := queryElem.Field(i)
								if fieldValue.CanInterface() {
									if db, ok := fieldValue.Interface().(*gorm.DB); ok && db != nil {
										return db, nil
									}
								}
							}
						}
					}
					// 尝试调用查询对象的方法
					queryMethods := []string{"Gorm", "DB", "GetDB", "GetGorm", "GetGormDB"}
					for _, methodName := range queryMethods {
						if method := queryElem.MethodByName(methodName); method.IsValid() {
							methodType := method.Type()
							if methodType.NumIn() == 0 && methodType.NumOut() > 0 {
								results := method.Call(nil)
								if len(results) > 0 {
									result := results[0]
									if result.Type().String() == "*gorm.DB" && !result.IsNil() {
										if db, ok := result.Interface().(*gorm.DB); ok && db != nil {
											return db, nil
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 尝试调用可能的方法（DB() 方法返回 *sql.DB，不是 *gorm.DB）
	methodNames := []string{"DB", "Gorm", "GetDB", "GetGorm", "GetGormDB"}
	for _, methodName := range methodNames {
		if method := ormValue.MethodByName(methodName); method.IsValid() {
			methodType := method.Type()
			if methodType.NumIn() == 0 {
				results := method.Call(nil)
				if len(results) > 0 {
					result := results[0]
					if result.Type().String() == "*gorm.DB" && !result.IsNil() {
						if db, ok := result.Interface().(*gorm.DB); ok && db != nil {
							return db, nil
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("无法从框架获取 GORM DB 实例")
}

// createGormConnection 根据配置直接创建 GORM 连接
func createGormConnection() (*gorm.DB, error) {
	config := facades.Config()
	connection := config.GetString("database.default", "sqlite")

	var db *gorm.DB
	var err error

	switch connection {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.GetString("database.connections.mysql.username", ""),
			config.GetString("database.connections.mysql.password", ""),
			config.GetString("database.connections.mysql.host", "127.0.0.1"),
			config.GetInt("database.connections.mysql.port", 3306),
			config.GetString("database.connections.mysql.database", "forge"),
			config.GetString("database.connections.mysql.charset", "utf8mb4"),
		)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.GetString("database.connections.postgres.host", "127.0.0.1"),
			config.GetInt("database.connections.postgres.port", 5432),
			config.GetString("database.connections.postgres.username", ""),
			config.GetString("database.connections.postgres.password", ""),
			config.GetString("database.connections.postgres.database", "forge"),
			config.GetString("database.connections.postgres.sslmode", "disable"),
		)
		if schema := config.GetString("database.connections.postgres.schema", ""); schema != "" {
			dsn += " search_path=" + schema
		}
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.GetString("database.connections.sqlserver.username", ""),
			config.GetString("database.connections.sqlserver.password", ""),
			config.GetString("database.connections.sqlserver.host", "127.0.0.1"),
			config.GetInt("database.connections.sqlserver.port", 1433),
			config.GetString("database.connections.sqlserver.database", "forge"),
		)
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	case "sqlite":
		dsn := config.GetString("database.connections.sqlite.database", "forge")
		db, err = gorm.Open(gormlite.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", connection)
	}

	if err != nil {
		errorlog.Record(context.Background(), "database", "创建 GORM 连接失败", map[string]any{
			"connection": connection,
			"error":      err.Error(),
		}, "创建 GORM 连接失败: %v", err)
		return nil, fmt.Errorf("创建 GORM 连接失败: %v", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(config.GetInt("database.pool.max_idle_conns", 10))
		sqlDB.SetMaxOpenConns(config.GetInt("database.pool.max_open_conns", 100))
		maxIdleTime := config.GetInt("database.pool.conn_max_idletime", 3600)
		maxLifetime := config.GetInt("database.pool.conn_max_lifetime", 3600)
		sqlDB.SetConnMaxIdleTime(time.Duration(maxIdleTime) * time.Second)
		sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	}

	facades.Log().Debugf("成功创建 GORM 连接（数据库类型: %s）", connection)
	return db, nil
}
