package utils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/goravel/framework/facades"
	"gorm.io/gorm"

	"goravel/app/utils/errorlog"
)

// GetGormDB 尝试获取原生 GORM DB 实例
// 通过反射从 Goravel ORM 获取底层 GORM DB
func GetGormDB() (*gorm.DB, error) {
	orm := facades.Orm()
	if orm == nil {
		errorlog.Record(context.Background(), "database", "ORM 未初始化", nil, "ORM 未初始化")
		return nil, fmt.Errorf("ORM 未初始化")
	}

	// 方法1: 尝试通过反射获取底层 GORM DB
	// Goravel 框架的 ORM 实现可能包含 *gorm.DB 字段
	ormValue := reflect.ValueOf(orm)
	if ormValue.Kind() == reflect.Ptr {
		ormValue = ormValue.Elem()
	}

	// 查找 *gorm.DB 类型的字段
	for i := 0; i < ormValue.NumField(); i++ {
		field := ormValue.Field(i)
		if field.Kind() == reflect.Ptr {
			if fieldType := field.Type(); fieldType.Elem().Name() == "DB" {
				// 检查是否是 gorm.DB
				if db, ok := field.Interface().(*gorm.DB); ok && db != nil {
					return db, nil
				}
			}
		}
		// 也尝试直接是 gorm.DB 类型
		if field.Kind() == reflect.Struct {
			if fieldType := field.Type(); fieldType.Name() == "DB" {
				if db, ok := field.Addr().Interface().(*gorm.DB); ok && db != nil {
					return db, nil
				}
			}
		}
	}

	// 方法2: 如果反射失败，尝试调用可能的方法
	// 检查是否有 GetDB 或 DB 方法
	methods := []string{"GetDB", "DB", "GormDB", "GetGormDB"}
	for _, methodName := range methods {
		method := ormValue.MethodByName(methodName)
		if method.IsValid() && method.Type().NumOut() > 0 {
			results := method.Call(nil)
			if len(results) > 0 {
				if db, ok := results[0].Interface().(*gorm.DB); ok && db != nil {
					return db, nil
				}
			}
		}
	}

	// 方法3: 尝试从接口类型断言
	if dbGetter, ok := orm.(interface{ GetDB() *gorm.DB }); ok {
		if db := dbGetter.GetDB(); db != nil {
			return db, nil
		}
	}

	err := fmt.Errorf("无法通过反射获取原生 GORM DB 实例，请检查 Goravel 框架实现或使用直接创建连接的方式")
	errorlog.Record(context.Background(), "database", "无法获取 GORM DB 实例", nil, "无法通过反射获取原生 GORM DB 实例，请检查 Goravel 框架实现或使用直接创建连接的方式")
	return nil, err
}
