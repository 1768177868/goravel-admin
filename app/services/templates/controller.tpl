package admin

import (
<<if .HasExport>>
<<if .ExportAsync>>
	"encoding/json"
<<end>>
	"fmt"
	"time"
<<end>>
	"strings"

	"github.com/goravel/framework/contracts/http"
<<if .HasExport>>
<<if .ExportAsync>>
	"github.com/goravel/framework/contracts/queue"
<<end>>
	"github.com/goravel/framework/facades"
<<if .ExportAsync>>
	"goravel/app/jobs"
	"goravel/app/models"
	"goravel/app/utils"
<<end>>
<<if not .ExportAsync>>
	"github.com/spf13/cast"
<<end>>
<<end>>

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	adminrequests "goravel/app/http/requests/admin"
	"goravel/app/http/response"
	"goravel/app/services"
)

type <<.ControllerName>> struct {}

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
	return services.Build<<.ModelName>>FiltersFromHTTP(ctx)
}

func New<<.ControllerName>>() *<<.ControllerName>> {
	return &<<.ControllerName>>{}
}

func (c *<<.ControllerName>>) <<.ServiceName>>(ctx http.Context) services.<<.ServiceName>> {
	return services.New<<.ServiceName>>(ctx)
}

// Index lists <<.ModelName>> records.
func (c *<<.ControllerName>>) Index(ctx http.Context) http.Response {
	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)

	filters := c.build<<.ModelName>>Filters(ctx)

	list, total, err := c.<<.ServiceName>>(ctx).GetList(filters, page, pageSize)
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

// Show returns <<.ModelName>> details.
func (c *<<.ControllerName>>) Show(ctx http.Context) http.Response {
	id := helpers.GetUintRoute(ctx, "id")
	item, err := c.<<.ServiceName>>(ctx).GetByID(id)
	if err != nil {
		return handleGeneratedServiceError(ctx, http.StatusNotFound, err)
	}

	return response.Success(ctx, http.Json{
		"<<.ModuleName>>": item,
	})
}

// Store creates a new <<.ModelName>>.
func (c *<<.ControllerName>>) Store(ctx http.Context) http.Response {
<<- if .HasCreate>>
	var req adminrequests.<<.RequestCreateName>>
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.<<.ServiceName>>(ctx).Create(&req)
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

// Update modifies an existing <<.ModelName>>.
func (c *<<.ControllerName>>) Update(ctx http.Context) http.Response {
<<- if .HasEdit>>
	id := helpers.GetUintRoute(ctx, "id")

	var req adminrequests.<<.RequestUpdateName>>
	if resp := validateGeneratedRequest(ctx, &req); resp != nil {
		return resp
	}

	item, err := c.<<.ServiceName>>(ctx).Update(id, &req)
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

// Destroy deletes a <<.ModelName>>.
func (c *<<.ControllerName>>) Destroy(ctx http.Context) http.Response {
<<- if .HasDelete>>
	id := helpers.GetUintRoute(ctx, "id")
	if err := c.<<.ServiceName>>(ctx).Delete(id); err != nil {
		return handleGeneratedServiceError(ctx, http.StatusInternalServerError, err)
	}

	return response.Success(ctx, "delete_success", http.Json{})
<<- else>>
	return response.Error(ctx, http.StatusForbidden, "delete_not_allowed")
<<end>>
}

// Export exports <<.ModelName>> records.
func (c *<<.ControllerName>>) Export(ctx http.Context) http.Response {
<<- if .HasExport>>
	adminID, err := helpers.GetAdminIDFromContext(ctx)
	if err != nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrUnauthorized.Code)
	}

	lockKey := fmt.Sprintf("export:<<.ModuleName>>s:lock:%d", adminID)
	lock := facades.Cache().Lock(lockKey, 10*time.Second)
	if !lock.Get() {
		return response.Error(ctx, http.StatusTooManyRequests, apperrors.ErrGetLockFailed.Code)
	}

	filters := c.build<<.ModelName>>Filters(ctx)
<<- if .ExportAsync>>
	filtersMap := utils.ExportFiltersToMap(filters)
	lang := utils.GetCurrentLanguage(ctx)
	timezone := helpers.GetCurrentTimezone(ctx)

	exportRecord := models.Export{
		AdminID: adminID,
		Type:    "<<.ModuleName>>s",
		Status:  models.ExportStatusProcessing,
		Disk:    "local",
		Path:    "",
	}
	if err := facades.Orm().Query().Create(&exportRecord); err != nil {
		return response.ErrorWithLog(ctx, "export", err)
	}

	exportArgsStruct := jobs.ExportArgs{
		ExportID: exportRecord.ID,
		AdminID:  adminID,
		Filters:  filtersMap,
		Type:     "<<.ModuleName>>s",
		Language: lang,
		Timezone: timezone,
	}

	exportArgsJSON, err := json.Marshal(exportArgsStruct)
	if err != nil {
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	exportArgs := []queue.Arg{
		{
			Type:  "string",
			Value: string(exportArgsJSON),
		},
	}

	if err := facades.Queue().Job(&jobs.Export<<.ModelName>>s{}, exportArgs).OnQueue("long-running").Dispatch(); err != nil {
		lock.Release()
		exportRecord.Status = models.ExportStatusFailed
		exportRecord.ErrorMsg = err.Error()
		facades.Orm().Query().Save(&exportRecord)
		return response.ErrorWithLog(ctx, "export", err)
	}

	return response.Success(ctx, http.Json{
		"export_id": exportRecord.ID,
		"message":   "queued",
	})
<<- else>>

	list, err := c.<<.ServiceName>>(ctx).GetAll<<.ModelName>>ForExport(filters)
	if err != nil {
		return response.ErrorWithLog(ctx, "<<.ModuleName>>", err, map[string]any{
			"action":   "export_<<.ModuleName>>s",
			"admin_id": adminID,
		})
	}

	headers := []string{
		<<- range .ListFields>>
		<<- if and .ShowInList (ne .Name "operation")>>
		"<<.Name>>",
		<<- end>>
		<<- end>>
	}

	timezone := helpers.GetCurrentTimezone(ctx)
	var data [][]string
	for _, row := range list {
		r := []string{
			<<- range .ListFields>>
			<<- if and .ShowInList (ne .Name "operation")>>
			<<- if eq .Name "created_at">>
			helpers.FormatCarbonWithTimezone(row.CreatedAt, timezone),
			<<- else if eq .Name "updated_at">>
			helpers.FormatCarbonWithTimezone(row.UpdatedAt, timezone),
			<<- else if eq .GoType "time.Time">>
			helpers.FormatTimeWithTimezone(row.<<.FieldName>>, timezone),
			<<- else if eq .GoType "string">>
			row.<<.FieldName>>,
			<<- else>>
			cast.ToString(row.<<.FieldName>>),
			<<- end>>
			<<- end>>
			<<- end>>
		}
		data = append(data, r)
	}

	ctx.WithValue("export_type", "<<.ModuleName>>s")

	return response.Export(ctx, "exported", headers, data, "<<.ModuleName>>s")
<<- end>>
<<- else>>
	return response.Error(ctx, http.StatusForbidden, "forbidden")
<<end>>
}

