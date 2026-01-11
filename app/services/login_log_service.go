package services

import (
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
)

type LoginLogService interface {
	// GetByID 根据ID获取登录日志
	GetByID(id uint, withAdmin bool) (*models.LoginLog, error)
	// GetList 获取登录日志列表
	GetList(filters LoginLogFilters, page, pageSize int) ([]models.LoginLog, int64, error)
}

// LoginLogFilters 登录日志查询过滤器
type LoginLogFilters struct {
	AdminID   string
	Username  string
	IP        string
	Status    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type LoginLogServiceImpl struct {
}

func NewLoginLogService() LoginLogService {
	return &LoginLogServiceImpl{}
}

// GetByID 根据ID获取登录日志
func (s *LoginLogServiceImpl) GetByID(id uint, withAdmin bool) (*models.LoginLog, error) {
	var log models.LoginLog
	query := facades.Orm().Query().Where("id", id)

	// 预加载关联
	if withAdmin {
		query = query.With("Admin")
	}

	if err := query.FirstOrFail(&log); err != nil {
		return nil, apperrors.ErrLogNotFound.WithError(err)
	}

	return &log, nil
}

// GetList 获取登录日志列表
func (s *LoginLogServiceImpl) GetList(filters LoginLogFilters, page, pageSize int) ([]models.LoginLog, int64, error) {
	query := facades.Orm().Query().Model(&models.LoginLog{})

	// 应用筛选条件
	if filters.AdminID != "" {
		query = query.Where("admin_id", filters.AdminID)
	}
	if filters.Username != "" {
		query = query.Where("username LIKE ?", "%"+filters.Username+"%")
	}
	if filters.IP != "" {
		query = query.Where("ip LIKE ?", "%"+filters.IP+"%")
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
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

	// 分页查询
	var logs []models.LoginLog
	var total int64
	if err := query.With("Admin").Paginate(page, pageSize, &logs, &total); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
