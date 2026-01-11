package services

import (
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
)

type DepartmentService interface {
	// GetByID 根据ID获取部门
	GetByID(id uint) (*models.Department, error)
	// GetList 获取部门列表
	GetList(filters DepartmentFilters, page, pageSize int) ([]models.Department, int64, error)
	// HasAdmins 检查部门是否有管理员
	HasAdmins(departmentID uint) (bool, error)
}

// DepartmentFilters 部门查询过滤器
type DepartmentFilters struct {
	Name      string
	Status    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type DepartmentServiceImpl struct {
	treeService TreeService
}

func NewDepartmentServiceImpl(treeService TreeService) *DepartmentServiceImpl {
	return &DepartmentServiceImpl{
		treeService: treeService,
	}
}

// GetByID 根据ID获取部门
func (s *DepartmentServiceImpl) GetByID(id uint) (*models.Department, error) {
	var department models.Department
	if err := facades.Orm().Query().Where("id", id).First(&department); err != nil {
		return nil, apperrors.ErrDepartmentNotFound.WithError(err)
	}
	return &department, nil
}

// GetList 获取部门列表
func (s *DepartmentServiceImpl) GetList(filters DepartmentFilters, page, pageSize int) ([]models.Department, int64, error) {
	query := facades.Orm().Query().Model(&models.Department{})

	// 应用筛选条件
	if filters.Name != "" {
		query = query.Where("name LIKE ?", "%"+filters.Name+"%")
	}
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
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
		orderBy = "sort:asc,id:asc"
	}
	query = helpers.ApplySort(query, orderBy, "sort:asc,id:asc")

	// 分页查询
	var departments []models.Department
	var total int64
	if err := query.Paginate(page, pageSize, &departments, &total); err != nil {
		return nil, 0, err
	}

	return departments, total, nil
}

// HasAdmins 检查部门是否有管理员
func (s *DepartmentServiceImpl) HasAdmins(departmentID uint) (bool, error) {
	count, err := facades.Orm().Query().Model(&models.Admin{}).Where("department_id", departmentID).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
