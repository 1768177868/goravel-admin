package utils

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/goravel/framework/facades"
	"gorm.io/gorm"
)

var (
	gormDBInstance *gorm.DB
	gormDBOnce     sync.Once
	gormDBErr      error
)

// GetGormDB 获取原生 GORM DB 实例
// 通过框架的 Query().Instance() 方法获取 GORM DB 实例
// 使用单例模式，确保只获取一次
func GetGormDB() (*gorm.DB, error) {
	gormDBOnce.Do(func() {
		gormDBInstance, gormDBErr = tryGetGormFromFramework()
		if gormDBErr == nil && gormDBInstance != nil {
			facades.Log().Infof("成功从框架获取 GORM DB 实例")
		}
	})
	return gormDBInstance, gormDBErr
}

// tryGetGormFromFramework 尝试从框架的 ORM 获取底层 GORM DB 实例
// 通过 Query().Instance() 方法获取 GORM DB 实例
func tryGetGormFromFramework() (*gorm.DB, error) {
	orm := facades.Orm()
	if orm == nil {
		return nil, fmt.Errorf("框架 ORM 不可用")
	}

	ormValue := reflect.ValueOf(orm)

	// 通过 Query() 方法获取查询对象，然后调用 Instance() 方法获取 GORM DB
	if queryMethod := ormValue.MethodByName("Query"); queryMethod.IsValid() {
		queryResults := queryMethod.Call(nil)
		if len(queryResults) > 0 {
			queryValue := queryResults[0]
			if queryValue.Kind() == reflect.Interface && !queryValue.IsNil() {
				if queryElem := queryValue.Elem(); queryElem.IsValid() {
					// 调用 Instance() 方法
					if instanceMethod := queryElem.MethodByName("Instance"); instanceMethod.IsValid() {
						methodType := instanceMethod.Type()
						if methodType.NumIn() == 0 && methodType.NumOut() > 0 {
							results := instanceMethod.Call(nil)
							if len(results) > 0 {
								result := results[0]
								// 检查返回类型是否为 *gorm.DB
								if result.Type().String() == "*gorm.DB" && !result.IsNil() {
									if db, ok := result.Interface().(*gorm.DB); ok && db != nil {
										facades.Log().Infof("成功通过 Instance() 方法从 Query 对象获取 GORM DB 实例")
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

	return nil, fmt.Errorf("无法从框架获取 GORM DB 实例")
}
