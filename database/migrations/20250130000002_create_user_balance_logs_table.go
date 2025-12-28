package migrations

type M20250130000002CreateUserBalanceLogsTable struct {
}

func (r *M20250130000002CreateUserBalanceLogsTable) Signature() string {
	return "20250130000002_create_user_balance_logs_table"
}

func (r *M20250130000002CreateUserBalanceLogsTable) Up() error {
	// 注意：使用 GORM Sharding 插件时，不需要手动创建分表
	// 插件会自动根据 ShardingKey 创建和管理分表
	// 这里只创建基础表结构（如果插件需要的话）
	// 实际上，GORM Sharding 会在首次插入时自动创建分表

	// 如果数据库需要基础表结构，可以创建（但通常不需要）
	// 因为 GORM Sharding 会自动管理分表

	return nil
}

func (r *M20250130000002CreateUserBalanceLogsTable) Down() error {
	// 删除所有分表（需要手动处理，因为 GORM Sharding 管理分表）
	// 这里不做任何操作，避免误删数据
	return nil
}
