package providers

import (
	"fmt"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/facades"
	"gorm.io/sharding"

	"goravel/app/utils"
	"goravel/database"
)

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
		return fmt.Errorf("获取 GORM DB 实例失败: %v", err)
	}

	// 配置 user_balance_logs 表的分表
	// ShardingKey: user_id - 分表键
	// NumberOfShards: 64 - 分表数量（建议为 2 的幂次）
	// PrimaryKeyGenerator: PKSnowflake - 主键生成器（确保全局唯一）
	err = db.Use(sharding.Register(sharding.Config{
		ShardingKey:         "user_id",
		NumberOfShards:      64,
		PrimaryKeyGenerator: sharding.PKSnowflake,
	}, "user_balance_logs"))
	if err != nil {
		return fmt.Errorf("注册 GORM Sharding 插件失败: %v", err)
	}

	facades.Log().Info("GORM Sharding 插件配置成功: user_balance_logs 表将按 user_id 分表（64个分表）")
	return nil
}
