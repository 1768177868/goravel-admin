package providers

import (
	"context"
	"fmt"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
	"gorm.io/sharding"

	"goravel/app/constants"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
	"goravel/database"
	"reflect"
)

// uintShardingAlgorithm 自定义分片算法函数，支持 uint 类型
// 根据 user_id 计算分表索引（0 到 UserBalanceLogsShards-1）
func uintShardingAlgorithm(value any) (suffix string, err error) {
	numberOfShards := constants.UserBalanceLogsShards
	if numberOfShards <= 0 {
		return "", fmt.Errorf("分片数量必须大于0,当前值：%d", numberOfShards)
	}

	val := reflect.ValueOf(value)
	var num int64
	switch val.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		num = int64(val.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		num = val.Int()
	default:
		return "", fmt.Errorf("不支持的 ShardingKey 类型: %T", value)
	}

	// 计算分片索引
	shardIndex := num % int64(numberOfShards)
	if shardIndex < 0 {
		shardIndex = -shardIndex
	}

	return fmt.Sprintf("_%d", shardIndex), nil
}

type DatabaseServiceProvider struct {
}

func (receiver *DatabaseServiceProvider) Register(app foundation.Application) {

}

func (receiver *DatabaseServiceProvider) Boot(app foundation.Application) {
	kernel := database.Kernel{}
	facades.Schema().Register(kernel.Migrations())
	facades.Seeder().Register(kernel.Seeders())

	// 配置 GORM Sharding 插件
	if err := receiver.initGormSharding(); err != nil {
		// 如果配置失败，记录日志但不阻止应用启动
		// 因为 GORM Sharding 是可选的，如果无法配置，数据会写入单表
		facades.Log().Warningf("GORM Sharding 配置失败: %v", err)
	}
}

// initGormSharding 初始化 GORM Sharding 插件
func (receiver *DatabaseServiceProvider) initGormSharding() error {
	// 尝试获取原生 GORM DB 实例
	db, err := utils.GetGormDB()
	if err != nil {
		errorlog.Record(context.Background(), "database", "获取 GORM DB 实例失败", nil, "获取 GORM DB 实例失败: %v", err)
		return fmt.Errorf("获取 GORM DB 实例失败: %v", err)
	}

	// 配置 user_balance_logs 表的分表
	// ShardingKey: user_id - 分表键（uint 类型）
	// NumberOfShards: 使用 constants.UserBalanceLogsShards 常量（建议为 2 的幂次）
	// ShardingAlgorithm: 自定义算法函数，支持 uint 类型
	// PrimaryKeyGenerator: PKSnowflake - 主键生成器（确保全局唯一）
	err = db.Use(sharding.Register(sharding.Config{
		ShardingKey:         "user_id",
		NumberOfShards:      constants.UserBalanceLogsShards,
		ShardingAlgorithm:   uintShardingAlgorithm,
		PrimaryKeyGenerator: sharding.PKSnowflake,
	}, "user_balance_logs"))
	if err != nil {
		errorlog.Record(context.Background(), "database", "注册 GORM Sharding 插件失败", map[string]any{
			"table": "user_balance_logs",
			"error": err.Error(),
		}, "注册 GORM Sharding 插件失败: %v", err)
		return fmt.Errorf("注册 GORM Sharding 插件失败: %v", err)
	}

	facades.Log().Infof("GORM Sharding 插件配置成功: user_balance_logs 表将按 user_id 分表（%d个分表）", constants.UserBalanceLogsShards)
	return nil
}
