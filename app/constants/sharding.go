package constants

// UserBalanceLogsShards 用户余额变动记录分表数量（已迁移至 config sharding.user_balance_logs_shards）。
// 保留常量供迁移/文档引用，运行时应使用 utils.GetUserBalanceLogsShards()。
const UserBalanceLogsShards = 4
