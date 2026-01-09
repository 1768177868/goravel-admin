package commands

import (
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
)

type OptimizeTables struct {
}

func (r *OptimizeTables) Signature() string {
	return "db:optimize-tables"
}

func (r *OptimizeTables) Description() string {

	// # MySQL: OPTIMIZE TABLE（整理碎片，回收空间）
	// # PostgreSQL: VACUUM (ANALYZE)（清理死元组并更新统计信息）
	//
	// # 传参方式：直接传表名（可多个）
	// go run . artisan db:optimize-tables payments
	// go run . artisan db:optimize-tables payments orders_202601 order_details_202601
	//
	// # 选项方式：--tables= 逗号分隔
	// go run . artisan db:optimize-tables --tables=payments,orders_202601
	//
	// # PostgreSQL 重度回收空间（风险高：会锁表，且耗时长）
	// go run . artisan db:optimize-tables payments --full=true
	//
	// # 帮助
	// go run . artisan db:optimize-tables --help

	return "优化表（MySQL: OPTIMIZE TABLE; PostgreSQL: VACUUM）"
}

func (r *OptimizeTables) Extend() command.Extend {
	return command.Extend{
		Category: "db",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "tables",
				Aliases: []string{"t"},
				Usage:   "要优化的表名列表（逗号分隔），也可以直接用参数方式传入多个表名",
			},
			&command.BoolFlag{
				Name:  "full",
				Value: false,
				Usage: "PostgreSQL 是否使用 VACUUM FULL（更重，可能锁表，默认 false）",
			},
		},
	}
}

func (r *OptimizeTables) Handle(ctx console.Context) error {
	dbConnection := strings.ToLower(facades.Config().GetString("database.default", "sqlite"))
	full := ctx.OptionBool("full")

	var tables []string
	// 1) 从 --tables 读取（逗号分隔）
	tablesFlag := strings.TrimSpace(ctx.Option("tables"))
	if tablesFlag != "" {
		parts := strings.Split(tablesFlag, ",")
		for _, p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				tables = append(tables, t)
			}
		}
	}
	// 2) 从参数读取（db:optimize-tables table1 table2 ...）
	for i := 0; ; i++ {
		arg := strings.TrimSpace(ctx.Argument(i))
		if arg == "" {
			break
		}
		tables = append(tables, arg)
	}

	if len(tables) == 0 {
		return fmt.Errorf("请提供要优化的表名，例如：go run . artisan db:optimize-tables payments 或使用 --tables=payments,orders_202601")
	}

	execOptimize := func(table string) error {
		var sql string
		switch dbConnection {
		case "mysql":
			sql = fmt.Sprintf("OPTIMIZE TABLE `%s`", table)
		case "postgres":
			if full {
				sql = fmt.Sprintf("VACUUM (FULL, ANALYZE) %s", table)
			} else {
				sql = fmt.Sprintf("VACUUM (ANALYZE) %s", table)
			}
		default:
			return fmt.Errorf("unsupported database: %s", dbConnection)
		}

		if _, err := facades.Orm().Query().Exec(sql); err != nil {
			return err
		}
		ctx.Info("✓ " + sql)
		return nil
	}

	ctx.Info("开始执行优化...")

	for _, table := range tables {
		if facades.Schema().HasTable(table) {
			if err := execOptimize(table); err != nil {
				return fmt.Errorf("optimize %s failed: %v", table, err)
			}
		}
	}

	ctx.Info("完成")
	return nil
}
