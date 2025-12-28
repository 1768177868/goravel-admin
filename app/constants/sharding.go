package constants

// UserBalanceLogsShards 用户余额变动记录表的分表数量
// 建议设置为 2 的幂次（如 4, 8, 16, 32, 64, 128 等）
// 修改此值即可，所有相关代码会自动使用此常量
const UserBalanceLogsShards = 4
