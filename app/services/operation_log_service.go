package services

import (
	"fmt"

	"github.com/goravel/framework/facades"

	"goravel/app/http/helpers"
	"goravel/app/models"
	"goravel/app/utils"
)

type OperationLogService interface {
	// GetByID 根据ID获取操作日志
	GetByID(id uint, withAdmin bool) (*models.OperationLog, error)
	// GetList 获取操作日志列表
	GetList(filters OperationLogFilters, page, pageSize int) ([]models.OperationLog, int64, error)
}

// OperationLogFilters 操作日志查询过滤器
type OperationLogFilters struct {
	AdminID   string
	Username  string
	Method    string
	Path      string
	Title     string
	IP        string
	Status    string
	Request   string
	StartTime string
	EndTime   string
	OrderBy   string
}

type OperationLogServiceImpl struct {
}

func NewOperationLogService() OperationLogService {
	return &OperationLogServiceImpl{}
}

// GetByID 根据ID获取操作日志
func (s *OperationLogServiceImpl) GetByID(id uint, withAdmin bool) (*models.OperationLog, error) {
	var log models.OperationLog
	query := facades.Orm().Query().Where("id", id)

	// 预加载关联
	if withAdmin {
		query = query.With("Admin")
	}

	if err := query.First(&log); err != nil {
		return nil, fmt.Errorf("操作日志不存在: %v", err)
	}

	return &log, nil
}

// GetList 获取操作日志列表
func (s *OperationLogServiceImpl) GetList(filters OperationLogFilters, page, pageSize int) ([]models.OperationLog, int64, error) {
	query := facades.Orm().Query().Model(&models.OperationLog{})

	// 应用筛选条件
	if filters.AdminID != "" {
		query = query.Where("admin_id", filters.AdminID)
	}
	if filters.Username != "" {
		// 通过用户名查找管理员ID
		var adminIDs []uint
		var admins []models.Admin
		if err := facades.Orm().Query().Where("username LIKE ?", "%"+filters.Username+"%").Get(&admins); err == nil {
			for _, admin := range admins {
				adminIDs = append(adminIDs, admin.ID)
			}
			if len(adminIDs) > 0 {
				idsAny := helpers.ConvertUintSliceToAny(adminIDs)
				query = query.WhereIn("admin_id", idsAny)
			} else {
				query = query.Where("admin_id", 0)
			}
		}
	}
	if filters.Method != "" {
		query = query.Where("method = ?", filters.Method)
	}
	if filters.Path != "" {
		query = query.Where("path LIKE ?", "%"+filters.Path+"%")
	}
	if filters.Title != "" {
		query = query.Where("title LIKE ?", "%"+filters.Title+"%")
	}
	if filters.IP != "" {
		query = query.Where("ip LIKE ?", "%"+filters.IP+"%")
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.Request != "" {
		// 使用工具函数应用全文索引搜索
		query = utils.ApplyFulltextSearch(query, "request", filters.Request)
	}
	if filters.StartTime != "" {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if filters.EndTime != "" {
		query = query.Where("created_at <= ?", filters.EndTime)
	}

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "id:desc"
	}
	query = helpers.ApplySort(query, orderBy, "id:desc")

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var logs []models.OperationLog
	err = query.With("Admin").
		Offset((page-1)*pageSize).
		Limit(pageSize).
		Find(&logs)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

