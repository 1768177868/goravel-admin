package admin

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
)

type DepartmentController struct {
	treeService       services.TreeService
	departmentService services.DepartmentService
}

func NewDepartmentController() *DepartmentController {
	treeService := services.NewTreeServiceImpl()
	return &DepartmentController{
		treeService:       treeService,
		departmentService: services.NewDepartmentServiceImpl(treeService),
	}
}

// Index 部门列表（树形结构）
func (r *DepartmentController) Index(ctx http.Context) http.Response {
	name := ctx.Request().Query("name", "")
	status := ctx.Request().Query("status", "")

	// 如果有搜索条件，返回扁平列表；否则返回树形结构
	if name != "" || status != "" {
		query := facades.Orm().Query().Model(&models.Department{})

		if name != "" {
			// 使用模型字段名，GORM 会自动转换为数据库字段名
			// 或者直接使用数据库字段名（根据迁移文件，字段名是 name）
			query = query.Where("name LIKE ?", "%"+name+"%")
		}
		if status != "" {
			query = query.Where("status", status)
		}

		var departments []models.Department
		if err := query.Order("sort asc, id asc").Get(&departments); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "query_failed")
		}

		return response.Success(ctx, "get_success", http.Json{
			"list": departments,
		})
	}

	// 无搜索条件时返回树形结构
	departments, err := r.treeService.BuildDepartmentTree(0)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Success(ctx, "get_success", http.Json{
		"list": departments,
	})
}

// Show 部门详情
func (r *DepartmentController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var department models.Department
	if err := facades.Orm().Query().Where("id", id).First(&department); err != nil {
		return response.Error(ctx, http.StatusNotFound, "department_not_found")
	}

	return response.Success(ctx, "get_success", http.Json{
		"department": department,
	})
}

// Store 创建部门
func (r *DepartmentController) Store(ctx http.Context) http.Response {
	parentID := cast.ToUint(ctx.Request().Input("parent_id", "0"))
	name := ctx.Request().Input("name")
	code := ctx.Request().Input("code")
	leader := ctx.Request().Input("leader")
	phone := ctx.Request().Input("phone")
	email := ctx.Request().Input("email")
	// 处理状态字段：需要正确处理 0 值
	allInputs := ctx.Request().All()
	var status uint8 = 1 // 默认启用
	if statusVal, exists := allInputs["status"]; exists {
		if statusVal != nil {
			status = cast.ToUint8(statusVal)
		}
	}
	sort := cast.ToInt(ctx.Request().Input("sort", "0"))
	remark := ctx.Request().Input("remark")

	if name == "" {
		return response.Error(ctx, http.StatusBadRequest, "department_name_required")
	}

	department := models.Department{
		ParentID: parentID,
		Name:     name,
		Code:     code,
		Leader:   leader,
		Phone:    phone,
		Email:    email,
		Status:   status,
		Sort:     sort,
		Remark:   remark,
	}

	if err := facades.Orm().Query().Create(&department); err != nil {
		facades.Log().Errorf("Create department error: %v, department data: %+v", err, department)
		return response.Error(ctx, http.StatusInternalServerError, "create_failed")
	}

	return response.Success(ctx, "create_success", http.Json{
		"department": department,
	})
}

// Update 更新部门
func (r *DepartmentController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var department models.Department
	if err := facades.Orm().Query().Where("id", id).First(&department); err != nil {
		return response.Error(ctx, http.StatusNotFound, "department_not_found")
	}

	parentID := ctx.Request().Input("parent_id", "")
	name := ctx.Request().Input("name")
	code := ctx.Request().Input("code")
	leader := ctx.Request().Input("leader")
	phone := ctx.Request().Input("phone")
	email := ctx.Request().Input("email")
	status := ctx.Request().Input("status", "")
	sort := ctx.Request().Input("sort", "")
	remark := ctx.Request().Input("remark")

	if name != "" {
		department.Name = name
	}
	if parentID != "" {
		department.ParentID = cast.ToUint(parentID)
	}
	if code != "" {
		department.Code = code
	}
	if leader != "" {
		department.Leader = leader
	}
	if phone != "" {
		department.Phone = phone
	}
	if email != "" {
		department.Email = email
	}
	if status != "" {
		department.Status = cast.ToUint8(status)
	}
	if sort != "" {
		department.Sort = cast.ToInt(sort)
	}
	if remark != "" {
		department.Remark = remark
	}

	if err := facades.Orm().Query().Save(&department); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "update_failed")
	}

	return response.Success(ctx, "update_success", http.Json{
		"department": department,
	})
}

// Destroy 删除部门
func (r *DepartmentController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	var department models.Department
	if err := facades.Orm().Query().Where("id", id).First(&department); err != nil {
		return response.Error(ctx, http.StatusNotFound, "department_not_found")
	}

	// 检查是否有子部门
	hasChildren, err := r.treeService.HasDepartmentChildren(id)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}
	if hasChildren {
		return response.Error(ctx, http.StatusBadRequest, "department_has_children")
	}

	// 检查是否有管理员
	hasAdmins, err := r.departmentService.HasAdmins(id)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}
	if hasAdmins {
		return response.Error(ctx, http.StatusBadRequest, "department_has_admins")
	}

	if _, err := facades.Orm().Query().Delete(&department); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "delete_failed")
	}

	return response.Success(ctx, "delete_success")
}
