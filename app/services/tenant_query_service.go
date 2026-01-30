package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
)

// TenantQueryService 租户查询服务（预留）
// 未启用多租户时：所有方法等价于直接使用 facades.Orm().Query()，不影响现有逻辑
// 启用多租户后：在中间件设置 ctx.WithValue("tenant_id", tenantID)，并为业务表添加 tenant_id 字段，
// 使用本服务的 Query/QueryModel/QueryTable 即可自动按租户过滤，无需在每个查询处手动 ScopeTenant
type TenantQueryService struct {
	ctx http.Context
}

// NewTenantQueryService 创建租户查询服务
// ctx 可为 nil，未启用多租户时行为与 facades.Orm().Query() 一致
func NewTenantQueryService(ctx http.Context) *TenantQueryService {
	return &TenantQueryService{ctx: ctx}
}

// Query 创建查询（预留：启用多租户后自动按 tenant_id 过滤）
func (s *TenantQueryService) Query() orm.Query {
	return helpers.ScopeTenant(s.ctx, facades.Orm().Query())
}

// QueryModel 创建模型查询（预留：启用多租户后自动按 tenant_id 过滤）
func (s *TenantQueryService) QueryModel(model any) orm.Query {
	return helpers.ScopeTenant(s.ctx, facades.Orm().Query().Model(model))
}

// QueryTable 创建表查询（预留：用于分表场景，启用多租户后自动按 tenant_id 过滤）
func (s *TenantQueryService) QueryTable(tableName string) orm.Query {
	return helpers.ScopeTenant(s.ctx, facades.Orm().Query().Table(tableName))
}

// QueryOrGlobal 创建查询并允许查询全局数据 tenant_id = 0（预留，适用于配置表等）
func (s *TenantQueryService) QueryOrGlobal() orm.Query {
	return helpers.ScopeTenantOrGlobal(s.ctx, facades.Orm().Query())
}

// QueryModelOrGlobal 创建模型查询并允许查询全局数据（预留）
func (s *TenantQueryService) QueryModelOrGlobal(model any) orm.Query {
	return helpers.ScopeTenantOrGlobal(s.ctx, facades.Orm().Query().Model(model))
}

// GetTenantID 获取当前请求的租户ID（预留）
func (s *TenantQueryService) GetTenantID() (uint, bool) {
	return helpers.GetTenantIDFromContext(s.ctx)
}

// IsSuperAdmin 是否为超级管理员（无租户ID）（预留）
func (s *TenantQueryService) IsSuperAdmin() bool {
	return helpers.IsSuperAdmin(s.ctx)
}
