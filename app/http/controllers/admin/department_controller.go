package admin

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
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

// findDepartmentByID 根据ID查找部门，如果不存在则返回错误响应
func (r *DepartmentController) findDepartmentByID(ctx http.Context, id uint) (*models.Department, http.Response) {
	return response.FindByID[models.Department](ctx, id, &response.FindByIDOptions{
		NotFoundMessageKey: "department_not_found",
	})
}

// buildQuery 构建部门查询
func (r *DepartmentController) buildQuery(ctx http.Context) orm.Query {
	name := ctx.Request().Query("name", "")
	status := ctx.Request().Query("status", "")
	// 使用辅助函数自动转换时区
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	query := facades.Orm().Query().Model(&models.Department{})

	if name != "" {
		// 使用模型字段名，GORM 会自动转换为数据库字段名
		// 或者直接使用数据库字段名（根据迁移文件，字段名是 name）
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		query = query.Where("status", status)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	orderBy := ctx.Request().Query("order_by", "")
	// 应用排序，默认排序为 sort asc, id asc
	query = helpers.ApplySort(query, orderBy, "sort:asc,id:asc")

	return query
}

// Index 部门列表（树形结构）
func (r *DepartmentController) Index(ctx http.Context) http.Response {
	name := ctx.Request().Query("name", "")
	status := ctx.Request().Query("status", "")
	// 使用辅助函数自动转换时区
	startTime := helpers.GetTimeQueryParam(ctx, "start_time")
	endTime := helpers.GetTimeQueryParam(ctx, "end_time")

	// 如果有搜索条件，返回扁平列表；否则返回树形结构
	if name != "" || status != "" || startTime != "" || endTime != "" {
		query := r.buildQuery(ctx)

		var departments []models.Department
		if err := query.Get(&departments); err != nil {
			return response.Error(ctx, http.StatusInternalServerError, "query_failed")
		}

		return response.Success(ctx, http.Json{
			"list": departments,
		})
	}

	// 无搜索条件时返回树形结构
	departments, err := r.treeService.BuildDepartmentTree(0)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, "query_failed")
	}

	return response.Success(ctx, http.Json{
		"list": departments,
	})
}

// Show 部门详情
func (r *DepartmentController) Show(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	department, resp := r.findDepartmentByID(ctx, id)
	if resp != nil {
		return resp
	}

	return response.Success(ctx, http.Json{
		"department": *department,
	})
}

// Store 创建部门
func (r *DepartmentController) Store(ctx http.Context) http.Response {
	// 使用请求验证
	var departmentCreate adminrequests.DepartmentCreate
	errors, err := ctx.Request().ValidateRequest(&departmentCreate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	now := carbon.Now()
	departmentData := map[string]any{
		"parent_id":  departmentCreate.ParentID,
		"name":       departmentCreate.Name,
		"code":       departmentCreate.Code,
		"leader":     departmentCreate.Leader,
		"phone":      departmentCreate.Phone,
		"email":      departmentCreate.Email,
		"status":     departmentCreate.Status,
		"sort":       departmentCreate.Sort,
		"remark":     departmentCreate.Remark,
		"created_at": now,
		"updated_at": now,
	}

	if err := facades.Orm().Query().Table("departments").Create(departmentData); err != nil {
		return response.ErrorWithLog(ctx, "department", err, map[string]any{
			"name": departmentCreate.Name,
		})
	}

	var department models.Department
	if err := facades.Orm().Query().Where("name", departmentCreate.Name).First(&department); err != nil {
		return response.ErrorWithLog(ctx, "department", err, map[string]any{
			"name": departmentCreate.Name,
		})
	}

	return response.Success(ctx, http.Json{
		"department": department,
	})
}

// Update 更新部门
func (r *DepartmentController) Update(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	department, resp := r.findDepartmentByID(ctx, id)
	if resp != nil {
		return resp
	}

	// 使用请求验证
	var departmentUpdate adminrequests.DepartmentUpdate
	errors, err := ctx.Request().ValidateRequest(&departmentUpdate)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if errors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", errors.All())
	}

	// 使用 All() 方法检查字段是否存在
	allInputs := ctx.Request().All()

	if _, exists := allInputs["name"]; exists {
		department.Name = departmentUpdate.Name
	}
	if _, exists := allInputs["parent_id"]; exists {
		department.ParentID = departmentUpdate.ParentID
	}
	if _, exists := allInputs["code"]; exists {
		department.Code = departmentUpdate.Code
	}
	if _, exists := allInputs["leader"]; exists {
		department.Leader = departmentUpdate.Leader
	}
	if _, exists := allInputs["phone"]; exists {
		department.Phone = departmentUpdate.Phone
	}
	if _, exists := allInputs["email"]; exists {
		department.Email = departmentUpdate.Email
	}
	if _, exists := allInputs["status"]; exists {
		department.Status = departmentUpdate.Status
	}
	if _, exists := allInputs["sort"]; exists {
		department.Sort = departmentUpdate.Sort
	}
	if _, exists := allInputs["remark"]; exists {
		department.Remark = departmentUpdate.Remark
	}

	if err := facades.Orm().Query().Save(department); err != nil {
		return response.ErrorWithLog(ctx, "department", err, map[string]any{
			"department_id": department.ID,
		})
	}

	return response.Success(ctx, http.Json{
		"department": *department,
	})
}

// Destroy 删除部门
func (r *DepartmentController) Destroy(ctx http.Context) http.Response {
	id := cast.ToUint(ctx.Request().Route("id"))
	department, resp := r.findDepartmentByID(ctx, id)
	if resp != nil {
		return resp
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

	if _, err := facades.Orm().Query().Delete(department); err != nil {
		return response.ErrorWithLog(ctx, "department", err, map[string]any{
			"department_id": department.ID,
		})
	}

	return response.Success(ctx)
}
