package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"

	"goravel/app/utils"
)

type AnalyzeStats struct {
}

func (r *AnalyzeStats) Signature() string {
	return "db:analyze-stats"
}

func (r *AnalyzeStats) Description() string {

	// # 默认：分析当前月+上个月的 orders/order_details 分表，并分析 payments 表
	// go run . artisan db:analyze-stats
	//
	// # 指定向前分析几个月（含当前月）
	// go run . artisan db:analyze-stats --months=2
	// go run . artisan db:analyze-stats --months=6
	//
	// # 指定某一个月（只分析该月的分表）
	// go run . artisan db:analyze-stats --month=202601
	//
	// # 只分析订单分表，不分析支付表
	// go run . artisan db:analyze-stats --payments=false
	//
	// # 只分析 payments 表
	// go run . artisan db:analyze-stats --orders=false --order-details=false
	//
	// # 帮助
	// go run . artisan db:analyze-stats --help

	return "更新订单分表与支付表统计信息（ANALYZE）"
}

func (r *AnalyzeStats) Extend() command.Extend {
	return command.Extend{
		Category: "db",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "month",
				Aliases: []string{"m"},
				Usage:   "指定月份(格式: YYYYMM,如:202512)，不指定则按 months 向前分析",
			},
			&command.IntFlag{
				Name:    "months",
				Aliases: []string{"n"},
				Value:   2,
				Usage:   "向前分析几个月（默认2：当前月+上个月）",
			},
			&command.BoolFlag{
				Name:  "orders",
				Value: true,
				Usage: "是否分析订单分表（orders_YYYYMM）",
			},
			&command.BoolFlag{
				Name:  "order-details",
				Value: true,
				Usage: "是否分析订单详情分表（order_details_YYYYMM）",
			},
			&command.BoolFlag{
				Name:  "payments",
				Value: true,
				Usage: "是否分析支付记录表（payments）",
			},
		},
	}
}

func (r *AnalyzeStats) Handle(ctx console.Context) error {
	dbConnection := strings.ToLower(facades.Config().GetString("database.default", "sqlite"))

	monthsFlag := ctx.OptionInt("months")
	if monthsFlag <= 0 {
		monthsFlag = 2
	}

	monthFlag := strings.TrimSpace(ctx.Option("month"))

	analyzeOrders := ctx.OptionBool("orders")
	analyzeOrderDetails := ctx.OptionBool("order-details")
	analyzePayments := ctx.OptionBool("payments")

	var months []time.Time
	if monthFlag != "" {
		parsedTime, err := time.ParseInLocation("200601", monthFlag, time.UTC)
		if err != nil {
			return fmt.Errorf("月份格式错误，应为 YYYYMM 格式（如:202512): %v", err)
		}
		months = []time.Time{parsedTime}
	} else {
		now := time.Now().UTC()
		currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < monthsFlag; i++ {
			months = append(months, currentMonth.AddDate(0, -i, 0))
		}
	}

	ctx.Info("开始执行 ANALYZE...")

	execAnalyze := func(table string) error {
		var sql string
		switch dbConnection {
		case "mysql":
			sql = fmt.Sprintf("ANALYZE TABLE `%s`", table)
		case "postgres":
			sql = fmt.Sprintf("ANALYZE %s", table)
		default:
			return fmt.Errorf("unsupported database: %s", dbConnection)
		}

		_, err := facades.Orm().Query().Exec(sql)
		if err != nil {
			return err
		}
		ctx.Info("✓ " + sql)
		return nil
	}

	if analyzeOrders {
		for _, m := range months {
			table := utils.GetShardingTableName("orders", m)
			if facades.Schema().HasTable(table) {
				if err := execAnalyze(table); err != nil {
					return fmt.Errorf("analyze %s failed: %v", table, err)
				}
			}
		}
	}

	if analyzeOrderDetails {
		for _, m := range months {
			table := utils.GetShardingTableName("order_details", m)
			if facades.Schema().HasTable(table) {
				if err := execAnalyze(table); err != nil {
					return fmt.Errorf("analyze %s failed: %v", table, err)
				}
			}
		}
	}

	if analyzePayments {
		if facades.Schema().HasTable("payments") {
			if err := execAnalyze("payments"); err != nil {
				return fmt.Errorf("analyze payments failed: %v", err)
			}
		}
	}

	ctx.Info("完成")
	return nil
}
