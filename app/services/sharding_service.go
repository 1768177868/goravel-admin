package services

import (
	"context"

	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/utils/errorlog"
	"goravel/database/migrations"
)

// TableCreator 表创建函数类型
type TableCreator func(tableName string) error

// ShardingService 分表服务接口
type ShardingService interface {
	// RegisterTableCreator 注册表的创建函数
	RegisterTableCreator(baseTableName string, creator TableCreator)
	// CreateShardingTable 创建分表
	CreateShardingTable(tableName, baseTableName string) error
	// EnsureShardingTable 确保分表存在，不存在则创建
	EnsureShardingTable(tableName, baseTableName string) error
}

type ShardingServiceImpl struct {
	ctx      context.Context
	creators map[string]TableCreator
}

func NewShardingService(ctx context.Context) ShardingService {
	service := &ShardingServiceImpl{
		ctx:      ctx,
		creators: make(map[string]TableCreator),
	}

	// 注册订单相关表的创建函数
	service.registerOrderTables()

	// 注册用户余额变动记录表的创建函数
	service.registerUserBalanceLogTables()

	return service
}

// registerOrderTables 注册订单表的创建函数
func (s *ShardingServiceImpl) registerOrderTables() {
	// 注册订单主表（调用 migrations 中的函数）
	s.RegisterTableCreator("orders", migrations.CreateOrdersShardingTable)

	// 注册订单详情表（调用 migrations 中的函数）
	s.RegisterTableCreator("order_details", migrations.CreateOrderDetailsShardingTable)
}

// registerUserBalanceLogTables 注册用户余额变动记录表的创建函数
func (s *ShardingServiceImpl) registerUserBalanceLogTables() {
	// 注册用户余额变动记录表（调用 migrations 中的函数）
	s.RegisterTableCreator("user_balance_logs", migrations.CreateUserBalanceLogsShardingTable)
}

// RegisterTableCreator 注册表的创建函数
func (s *ShardingServiceImpl) RegisterTableCreator(baseTableName string, creator TableCreator) {
	s.creators[baseTableName] = creator
}

// CreateShardingTable 创建分表
func (s *ShardingServiceImpl) CreateShardingTable(tableName, baseTableName string) error {
	creator, exists := s.creators[baseTableName]
	if !exists {
		return apperrors.ErrBaseTableNotRegistered.WithParams(map[string]any{
			"base_table_name": baseTableName,
		})
	}

	return creator(tableName)
}

// EnsureShardingTable 确保分表存在，不存在则创建
func (s *ShardingServiceImpl) EnsureShardingTable(tableName, baseTableName string) error {
	// 检查表是否已存在
	if facades.Schema().HasTable(tableName) {
		return nil
	}

	// 创建表
	if err := s.CreateShardingTable(tableName, baseTableName); err != nil {
		errorlog.Record(s.ctx, "sharding", "创建分表失败", map[string]any{
			"table_name":      tableName,
			"base_table_name": baseTableName,
			"error":           err.Error(),
		}, "创建分表 %s 失败: %v", tableName, err)
		return apperrors.ErrCreateShardingTableFailed.WithError(err).WithParams(map[string]any{
			"table_name": tableName,
		})
	}

	facades.Log().Infof("自动创建分表: %s", tableName)
	return nil
}
