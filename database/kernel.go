package database

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"

	"goravel/database/migrations"
	"goravel/database/seeders"
)

type Kernel struct {
}

func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateUsersTable{},
		&migrations.M20210101000002CreateJobsTable{},
		// 后台管理系统相关表
		&migrations.M20250101000001CreateDepartmentsTable{},
		&migrations.M20250101000002CreateAdminsTable{},
		&migrations.M20250101000003CreateRolesTable{},
		&migrations.M20250101000004CreatePermissionsTable{},
		&migrations.M20250101000005CreateMenusTable{},
		&migrations.M20250101000006CreateDictionariesTable{},
		&migrations.M20250101000007CreateAdminRoleTable{},
		&migrations.M20250101000008CreateRolePermissionTable{},
		&migrations.M20250101000009CreateRoleMenuTable{},
		&migrations.M20250101000010CreateOperationLogsTable{},
		&migrations.M20250101000011CreateLoginLogsTable{},
		&migrations.M20250101000012CreateSystemLogsTable{},
		&migrations.M20250101000014CreatePersonalAccessTokensTable{},
		// 以下迁移已不需要，因为 users 表创建时已包含所有字段
		// &migrations.M20250330911908AddColumnsToUsersTable{},
		// &migrations.M20250331093125AlertColumnsOfUsersTable{},
		// &migrations.M20250101000015AddUsernamePasswordToUsersTable{},
	}
}

func (kernel Kernel) Seeders() []seeder.Seeder {
	return []seeder.Seeder{
		&seeders.DatabaseSeeder{},
		&seeders.AdminSeeder{},
	}
}
