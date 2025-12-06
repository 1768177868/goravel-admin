package jobs

import (
	"time"

	"github.com/goravel/framework/facades"

	"goravel/app/errors"
)

// GenerateReportArgs 生成报表任务的参数结构体
type GenerateReportArgs struct {
	StartDate string `validate:"required"`
	EndDate   string `validate:"required"`
}

// GenerateReport 生成报表任务
type GenerateReport struct {
}

func (r *GenerateReport) Signature() string {
	return "generate_report"
}

// Handle 处理生成报表任务
// 
// 参数:
//   - args[0]: GenerateReportArgs 结构体或 map[string]any
//   - args[1]: 如果 args[0] 是 string，则 args[1] 是结束日期
//
// 返回:
//   - error: 错误信息
func (r *GenerateReport) Handle(args ...any) error {
	var startDate, endDate string

	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing report arguments")
	}

	switch v := args[0].(type) {
	case GenerateReportArgs:
		startDate = v.StartDate
		endDate = v.EndDate
	case map[string]any:
		if sd, ok := v["start_date"].(string); ok {
			startDate = sd
		}
		if ed, ok := v["end_date"].(string); ok {
			endDate = ed
		}
	case string:
		// 兼容旧版本：按位置解析
		startDate = v
		if len(args) >= 2 {
			if ed, ok := args[1].(string); ok {
				endDate = ed
			} else {
				return errors.ErrInvalidArgument.WithMessage("invalid end date type")
			}
		} else {
			return errors.ErrInvalidArgument.WithMessage("missing end date")
		}
	default:
		return errors.ErrInvalidArgument.WithMessage("invalid argument type")
	}

	if startDate == "" {
		return errors.ErrInvalidArgument.WithMessage("start date is required")
	}
	if endDate == "" {
		return errors.ErrInvalidArgument.WithMessage("end date is required")
	}

	// 验证日期格式
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return errors.ErrInvalidArgument.WithMessage("invalid start date format, expected YYYY-MM-DD")
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return errors.ErrInvalidArgument.WithMessage("invalid end date format, expected YYYY-MM-DD")
	}

	facades.Log().Infof("📊 [Job] 生成报表 - 开始日期: %s, 结束日期: %s", startDate, endDate)
	// 实际场景中这里会查询数据、生成Excel报表等
	return nil
}
