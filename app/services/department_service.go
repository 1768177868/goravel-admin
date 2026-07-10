package services

import (
	"context"
	appfacades "goravel/app/facades"

	"github.com/dromara/carbon/v2"

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
	// Create 创建部门
	Create(parentID uint, name, code, leader, phone, email, remark string, status uint8, sort int) (*models.Department, error)
	// Update 更新部门
	Update(department *models.Department) error
	// Delete 删除部门
	Delete(department *models.Department) error
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
	ctx         context.Context
	treeService TreeService
}

func NewDepartmentServiceImpl(ctx context.Context, treeService TreeService) *DepartmentServiceImpl {
	return &DepartmentServiceImpl{
		ctx:         ctx,
		treeService: treeService,
	}
}

// GetByID 根据ID获取部门
func (s *DepartmentServiceImpl) GetByID(id uint) (*models.Department, error) {
	var department models.Department
	if err := appfacades.OrmQuery(s.ctx).Where("id", id).FirstOrFail(&department); err != nil {
		return nil, apperrors.ErrDepartmentNotFound.WithError(err)
	}
	return &department, nil
}

// GetList 获取部门列表
func (s *DepartmentServiceImpl) GetList(filters DepartmentFilters, page, pageSize int) ([]models.Department, int64, error) {
	query := appfacades.OrmQuery(s.ctx).Model(&models.Department{})

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
	count, err := appfacades.OrmQuery(s.ctx).Model(&models.Admin{}).Where("department_id", departmentID).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Create 创建部门
func (s *DepartmentServiceImpl) Create(parentID uint, name, code, leader, phone, email, remark string, status uint8, sort int) (*models.Department, error) {
	department := &models.Department{}
	createData := map[string]any{
		"parent_id":  parentID,
		"name":       name,
		"code":       code,
		"leader":     leader,
		"phone":      phone,
		"email":      email,
		"remark":     remark,
		"status":     status,
		"sort":       sort,
		"created_at": carbon.Now(),
		"updated_at": carbon.Now(),
	}

	if err := appfacades.OrmQuery(s.ctx).Model(department).Create(createData); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return department, nil
}

// Update 更新部门
func (s *DepartmentServiceImpl) Update(department *models.Department) error {
	if err := appfacades.OrmQuery(s.ctx).Save(department); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}
	return nil
}

// Delete 删除部门
func (s *DepartmentServiceImpl) Delete(department *models.Department) error {
	if _, err := appfacades.OrmQuery(s.ctx).Delete(department); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}
	return nil
}
