package admin

import (
<<if .HasExport>>
	"encoding/json"
	"fmt"
<<end>>
	"strings"

<<if .HasExport>>
	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
<<end>>
	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
<<if .HasExport>>
	"goravel/app/jobs"
	"goravel/app/models"
<<end>>
	"goravel/app/services"
<<if .HasExport>>
	"goravel/app/utils"
<<end>>
)

type <<.ControllerName>> struct {
	<<.ServiceName>> services.<<.ServiceName>>
}

func handleGeneratedServiceError(ctx http.Context, status int, err error) http.Response {
	if businessErr, ok := apperrors.GetBusinessError(err); ok {
		if businessErr.Code == "params_error" || businessErr.Code == "invalid_argument" {
			return response.Error(ctx, http.StatusBadRequest, businessErr.Code)
		}
		if businessErr.Code == "record_not_found" || strings.HasSuffix(businessErr.Code, "_not_found") {
			return response.Error(ctx, http.StatusNotFound, businessErr.Code)
		}
		return response.Error(ctx, status, businessErr.Code)
	}
	return response.Error(ctx, status, err.Error())
}

func validateGeneratedRequest(ctx http.Context, req http.FormRequest) http.Response {
	validationErrors, err := ctx.Request().ValidateRequest(req)
	if err != nil {
		return response.Error(ctx, http.StatusBadRequest, err.Error())
	}
	if validationErrors != nil {
		return response.ValidationError(ctx, http.StatusBadRequest, "validation_failed", validationErrors.All())
	}
	return nil
}

func (c *<<.ControllerName>>) build<<.ModelName>>Filters(ctx http.Context) services.<<.ModelName>>Filters {
	return services.<<.ModelName>>Filters{
<<- range .SearchableFields>>
		<<.PascalName>>: ctx.Request().Query("<<.Name>>", ""),
		<<- if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>
		<<.PascalName>>Start: ctx.Request().Query("<<.Name>>_start", ""),
		<<.PascalName>>End: ctx.Request().Query("<<.Name>>_end", ""),
		<<- end>>
<<- end>>
	}
}

func New<<.ControllerName>>() *<<.ControllerName>> {
	return &<<.ControllerName>>{
		<<.ServiceName>>: services.New<<.ServiceName>>(),
	}
}

// Index <<.ModelName>>列表
func (c *<<.ControllerName>>) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page",1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := c.build<<.ModelName>>Filters(ctx)

	list, total, err := c.<<.ServiceName>>.GetList(filters, page, pageSize)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// Show <<.ModelName>>详情
func (c *<<.ControllerName>>) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	item, err := c.<<.ServiceName>>.GetByID(id)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusNotFound, err)
	}

	return response.Success(ctx, http.Json{
		"<<.ModuleName>>": item,
	})
}

// Store 创建<<.ModelName>>
func (c *<<.ControllerName>>) Store(ctx http.Context) http.Response {
<<- if .HasCreate>>
	var req adminrequests.<<.RequestCreateName>>
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.<<.ServiceName>>.Create(&req)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"<<.ModuleName>>": item,
	})
<<- else>>
	return response.Error(ctx, http.StatusForbidden, "create_not_allowed")
<<end>>
}

// Update 更新<<.ModelName>>
func (c *<<.ControllerName>>) Update(ctx http.Context) http.Response {
<<- if .HasEdit>>
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.<<.RequestUpdateName>>
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.<<.ServiceName>>.Update(id, &req)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, http.Json{
		"<<.ModuleName>>": item,
	})
<<- else>>
	return response.Error(ctx, http.StatusForbidden, "update_not_allowed")
<<end>>
}

// Destroy 删除<<.ModelName>>
func (c *<<.ControllerName>>) Destroy(ctx http.Context) http.Response {
<<- if .HasDelete>>
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.<<.ServiceName>>.Delete(id); err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "delete_success", http.Json{})
<<- else>>
	return response.Error(ctx, http.StatusForbidden, "delete_not_allowed")
<<end>>
}

// Export 导出<<.ModelName>>
func (c *<<.ControllerName>>) Export(ctx http.Context) http.Response {
<<- if .HasExport>>
	filters := c.build<<.ModelName>>Filters(ctx)
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, "unauthorized")
	}

	lockKey := fmt.Sprintf("export:<<.ModuleName>>s:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10)
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, "export_in_progress")
	}

	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = utils.GetConfigValue("storage", "export_disk", "")
	}
	if disk == "" {
		disk = "local"
	}

	exportRecord := models.Export{
		AdminID: adminID,
		Type:    "<<.ModuleName>>s",
		Status:  models.ExportStatusProcessing,
		Disk:    disk,
		Path:    "",
	}
	if err := facades.Orm().Query().Create(&exportRecord); err != nil {
		lock.Release()
		return response.ErrorWithLog(ctx, "export", err)
	}

	filtersMap := map[string]any{}
	<<- range .SearchableFields>>
	if filters.<<.PascalName>> != "" {
		filtersMap["<<.Name>>"] = filters.<<.PascalName>>
	}
	<<- if or (eq .SearchUIType "daterange") (eq .SearchUIType "datetimerange")>>
	if filters.<<.PascalName>>Start != "" {
		filtersMap["<<.Name>>_start"] = filters.<<.PascalName>>Start
	}
	if filters.<<.PascalName>>End != "" {
		filtersMap["<<.Name>>_end"] = filters.<<.PascalName>>End
	}
	<<- end>>
	<<- end>>

	exportArgsStruct := jobs.ExportGenericArgs{
		ExportArgs: jobs.ExportArgs{
			ExportID: exportRecord.ID,
			AdminID:  adminID,
			Filters:  filtersMap,
			Type:     "<<.ModuleName>>s",
			Language: utils.GetCurrentLanguage(ctx),
			Timezone: helpers.GetCurrentTimezone(ctx),
		},
		Table:      "<<.TableName>>",
		FilePrefix: "<<.ModuleName>>s",
		HeaderKeys: []string{
			<<- range .ListFields>>
			<<- if and .ShowInList (ne .Name "operation")>>
			"<<.Name>>",
			<<- end>>
			<<- end>>
		},
		Columns: []string{
			<<- range .ListFields>>
			<<- if and .ShowInList (ne .Name "operation")>>
			"<<.Name>>",
			<<- end>>
			<<- end>>
		},
		SearchTypes: map[string]string{
			<<- range .SearchableFields>>
			"<<.Name>>": "<<.SearchType>>",
			<<- end>>
		},
	}

	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		lock.Release()
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		_ = facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.ExportGeneric{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		_ = facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	lock.Release()
	exportID := exportRecord.ID

	return response.Success(ctx, "export_task_submitted", http.Json{
		"export_id": exportID,
	})
<<- else>>
	return response.Error(ctx, http.StatusForbidden, "export_not_allowed")
<<end>>
}
