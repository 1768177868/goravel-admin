package services

import (
	"github.com/goravel/framework/facades"

	"goravel/app/models"
)

type DepartmentService interface {
	// HasAdmins 检查部门是否有管理员
	HasAdmins(departmentID uint) (bool, error)
}

type DepartmentServiceImpl struct {
	treeService TreeService
}

func NewDepartmentServiceImpl(treeService TreeService) *DepartmentServiceImpl {
	return &DepartmentServiceImpl{
		treeService: treeService,
	}
}

// HasAdmins 检查部门是否有管理员
func (s *DepartmentServiceImpl) HasAdmins(departmentID uint) (bool, error) {
	count, err := facades.Orm().Query().Model(&models.Admin{}).Where("department_id", departmentID).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
