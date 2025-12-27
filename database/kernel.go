package database

import (
	"goravel/database/migrations"
	"goravel/database/seeders"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
)

type Kernel struct {
}

func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000002CreateJobsTable{},
		// 后台管理系统相关表
		&migrations.M20250101000001CreateDepartmentsTable{},
		&migrations.M20250101000002CreateAdminsTable{},
		&migrations.M20250101000003CreateRolesTable{},
		&migrations.M20250101000004CreatePermissionsTable{},
		&migrations.M20250101000005CreateMenusTable{},
		&migrations.M20250101000006CreateDictionariesTable{},
		&migrations.M20250101000015CreateConfigsTable{},
		&migrations.M20250101000016CreateBlacklistsTable{},
		&migrations.M20250101000007CreateAdminRoleTable{},
		&migrations.M20250101000008CreateRolePermissionTable{},
		&migrations.M20250101000009CreateRoleMenuTable{},
		&migrations.M20250101000010CreateOperationLogsTable{},
		&migrations.M20250101000018AddTitleToOperationLogs{},
		&migrations.M20250101000011CreateLoginLogsTable{},
		&migrations.M20250101000012CreateSystemLogsTable{},
		&migrations.M20250201000016AddTraceIdToSystemLogsTable{},
		&migrations.M20250101000014CreatePersonalAccessTokensTable{},
		&migrations.M20250101000017AddOnlineUserFieldsToPersonalAccessTokens{},
		&migrations.M20250201000003CreateNotificationsTable{},
		&migrations.M20250301000021CreateExportsTable{},
		&migrations.M20250301000022CreateAttachmentsTable{},
		&migrations.M20250301000023AddDisplayNameToAttachments{},
		&migrations.M20250101000024AddGoogleSecretToAdmins{},
		&migrations.M20250101000025AddLinkTypeToMenus{},
		&migrations.M20250101000026ModifyMenusPathLength{},
		&migrations.M20251227063517AddFulltextIndexToOperationLogsRequest{},
	}
}
func (kernel Kernel) Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
		&seeders.MenuSeeder{},       // 菜单（需要先创建，因为权限依赖）
		&seeders.PermissionSeeder{}, // 权限（依赖菜单）
		&seeders.AdminSeeder{},      // 管理员、部门、角色（最后执行，关联权限和菜单）
		&seeders.DictionarySeeder{}, // 字典数据
	}
}
