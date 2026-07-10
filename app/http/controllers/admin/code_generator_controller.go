package admin

import (
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/response"
	"goravel/app/services"
)

type CodeGeneratorController struct {}


type GenerateRequest struct {
	ModuleName string                 `json:"module_name"`
	TableName  string                 `json:"table_name"`
	Fields     []services.FieldConfig `json:"fields"`
	Files      []string               `json:"files"`
	Options    map[string]bool        `json:"options"`
}

type PreviewRequest struct {
	ModuleName string                 `json:"module_name"`
	TableName  string                 `json:"table_name"`
	Fields     []services.FieldConfig `json:"fields"`
	FileType   string                 `json:"file_type"`
	Options    map[string]bool        `json:"options"`
}

type SaveRequest struct {
	ModuleName string                 `json:"module_name"`
	TableName  string                 `json:"table_name"`
	Fields     []services.FieldConfig `json:"fields"`
	Force      bool                   `json:"force"`
	Files      []string               `json:"files"`
	Options    map[string]bool        `json:"options"`
}

type GenerateWithAIRequest struct {
	Description string `json:"description"`
}

func NewCodeGeneratorController() *CodeGeneratorController {
	return &CodeGeneratorController{}
}

func (c *CodeGeneratorController) codeGeneratorService(ctx http.Context) services.CodeGeneratorService {
	return services.NewCodeGeneratorService(ctx)
}


// Generate 生成CRUD代码
func (c *CodeGeneratorController) Generate(ctx http.Context) http.Response {
	var req GenerateRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_fields")
	}

	if req.ModuleName == "" {
		return response.Error(ctx, http.StatusBadRequest, "module_name_required")
	}
	if req.TableName == "" {
		return response.Error(ctx, http.StatusBadRequest, "table_name_required")
	}

	files, err := c.codeGeneratorService(ctx).Generate(req.ModuleName, req.TableName, req.Fields, req.Files, req.Options)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"files": files,
	})
}

// Preview 预览生成的代码
func (c *CodeGeneratorController) Preview(ctx http.Context) http.Response {
	var req PreviewRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_fields")
	}

	if req.ModuleName == "" {
		return response.Error(ctx, http.StatusBadRequest, "module_name_required")
	}
	if req.TableName == "" {
		return response.Error(ctx, http.StatusBadRequest, "table_name_required")
	}
	if req.FileType == "" {
		return response.Error(ctx, http.StatusBadRequest, "file_type_required")
	}

	code, err := c.codeGeneratorService(ctx).Preview(req.ModuleName, req.TableName, req.Fields, req.FileType, req.Options)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"code": code,
	})
}

// Save 保存生成的代码到文件系统
func (c *CodeGeneratorController) Save(ctx http.Context) http.Response {
	var req SaveRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_fields")
	}

	if req.ModuleName == "" {
		return response.Error(ctx, http.StatusBadRequest, "module_name_required")
	}
	if req.TableName == "" {
		return response.Error(ctx, http.StatusBadRequest, "table_name_required")
	}

	var savedFiles []string
	var err error

	if req.Force {
		savedFiles, err = c.codeGeneratorService(ctx).ForceSave(req.ModuleName, req.TableName, req.Fields, req.Files, req.Options)
	} else {
		savedFiles, err = c.codeGeneratorService(ctx).Save(req.ModuleName, req.TableName, req.Fields, req.Files, req.Options)
	}

	if err != nil {
		if filesExistErr, ok := err.(*services.FilesExistError); ok {
			return ctx.Response().Json(409, http.Json{
				"code":       409,
				"message":    facades.Lang(ctx).Get("files_exist"),
				"error_code": "files_exist",
				"files":      filesExistErr.Files,
			})
		}
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"saved_files": savedFiles,
	})
}

// GetFieldTypes 获取支持的字段类型
func (c *CodeGeneratorController) GetFieldTypes(ctx http.Context) http.Response {
	fieldTypes := c.codeGeneratorService(ctx).GetFieldTypes()
	return response.Success(ctx, http.Json{
		"field_types": fieldTypes,
	})
}

// GetTables 获取数据库表列表
func (c *CodeGeneratorController) GetTables(ctx http.Context) http.Response {
	tables, err := c.codeGeneratorService(ctx).GetTables()
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}
	return response.Success(ctx, http.Json{
		"tables": tables,
	})
}

// GetTableColumns 获取表字段
func (c *CodeGeneratorController) GetTableColumns(ctx http.Context) http.Response {
	tableName := ctx.Request().Query("table_name")
	if tableName == "" {
		return response.Error(ctx, http.StatusBadRequest, "table_name_required")
	}

	fields, err := c.codeGeneratorService(ctx).GetTableColumns(tableName)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}
	return response.Success(ctx, http.Json{
		"fields": fields,
	})
}

// GenerateWithAI 使用 AI 生成模块配置
func (c *CodeGeneratorController) GenerateWithAI(ctx http.Context) http.Response {
	var req GenerateWithAIRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return response.Error(ctx, http.StatusBadRequest, "invalid_fields")
	}

	if req.Description == "" {
		return response.Error(ctx, http.StatusBadRequest, "description_required")
	}

	config, err := c.codeGeneratorService(ctx).GenerateWithAI(ctx, req.Description)
	if err != nil {
		// 检查是否是 JSON 解析错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "parse AI response") || strings.Contains(errMsg, "parse") || strings.Contains(errMsg, "JSON") {
			// JSON 解析错误返回 400，并返回友好的错误信息
			return ctx.Response().Json(400, http.Json{
				"code":       400,
				"message":    errMsg,
				"error_code": "ai_response_parse_error",
			})
		}
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusInternalServerError, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, err.Error())
	}

	return response.Success(ctx, http.Json{
		"config": config,
	})
}
