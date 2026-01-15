package services

import (
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type FieldConfig struct {
	Name         string          `json:"name"`
	Label        string          `json:"label"`
	GoType       string          `json:"go_type"`
	DBType       string          `json:"db_type"`
	Validators   []string        `json:"validators"`
	Required     bool            `json:"required"`
	Searchable   bool            `json:"searchable"`
	Sortable     bool            `json:"sortable"`
	SearchType   string          `json:"search_type"`
	SearchUIType string          `json:"search_ui_type"`
	Dictionary   string          `json:"dictionary"`
	ApiUrl       string          `json:"api_url"`
	Relation     *RelationConfig `json:"relation"`
}

type RelationConfig struct {
	Table        string `json:"table"`         // 关联表名
	ForeignKey   string `json:"foreign_key"`   // 外键字段
	DisplayField string `json:"display_field"` // 显示字段
	RelationType string `json:"relation_type"` // 关联类型：hasOne, belongsTo, hasMany
}

type FieldType struct {
	Label      string   `json:"label"`
	Value      string   `json:"value"`
	GoType     string   `json:"go_type"`
	DBType     string   `json:"db_type"`
	Validators []string `json:"validators"`
}

type GeneratedFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type FilesExistError struct {
	Files []string
}

func (e *FilesExistError) Error() string {
	return fmt.Sprintf("files already exist: %v", e.Files)
}

type CodeGeneratorService interface {
	Preview(moduleName, tableName string, fields []FieldConfig, fileType string, options map[string]bool) (string, error)
	Save(moduleName, tableName string, fields []FieldConfig, selectedFiles []string, options map[string]bool) ([]string, error)
	ForceSave(moduleName, tableName string, fields []FieldConfig, selectedFiles []string, options map[string]bool) ([]string, error)
	GetFieldTypes() []FieldType
	Generate(moduleName, tableName string, fields []FieldConfig, selectedFiles []string, options map[string]bool) ([]GeneratedFile, error)
}

type CodeGeneratorServiceImpl struct{}

//go:embed templates/*
var templates embed.FS

func NewCodeGeneratorService() CodeGeneratorService {
	return &CodeGeneratorServiceImpl{}
}

func (s *CodeGeneratorServiceImpl) Generate(moduleName, tableName string, fields []FieldConfig, selectedFiles []string, options map[string]bool) ([]GeneratedFile, error) {
	var files []GeneratedFile

	generators := []struct {
		fileType string
		generate func(string, string, []FieldConfig, map[string]bool) (GeneratedFile, error)
	}{
		{"model", s.generateModel},
		{"controller", s.generateController},
		{"service", s.generateService},
		{"request_create", s.generateRequestCreate},
		{"request_update", s.generateRequestUpdate},
		{"migration", s.generateMigration},
		{"api", s.generateFrontendAPI},
		{"list_page", s.generateFrontendListPage},
		{"form_page", s.generateFrontendFormPage},
	}

	// Create a map for faster lookup of selected files
	selectedMap := make(map[string]bool)
	if len(selectedFiles) > 0 {
		for _, f := range selectedFiles {
			selectedMap[f] = true
		}
	}

	for _, gen := range generators {
		// If selectedFiles is provided, only generate selected files
		if len(selectedFiles) > 0 && !selectedMap[gen.fileType] {
			continue
		}

		file, err := gen.generate(moduleName, tableName, fields, options)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s: %w", gen.fileType, err)
		}

		if strings.HasSuffix(file.Path, ".go") {
			formatted, err := format.Source([]byte(file.Content))
			if err == nil {
				file.Content = string(formatted)
			}
		}

		files = append(files, file)
	}

	return files, nil
}

func (s *CodeGeneratorServiceImpl) Preview(moduleName, tableName string, fields []FieldConfig, fileType string, options map[string]bool) (string, error) {
	templateName, err := s.getTemplateName(fileType)
	if err != nil {
		return "", err
	}

	templateContent, err := templates.ReadFile(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("preview").Delims("<<", ">>").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := s.buildTemplateData(moduleName, tableName, fields, fileType, options)
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	content := builder.String()
	if strings.Contains(templateName, ".go") || fileType == "model" || fileType == "controller" || fileType == "service" || fileType == "request_create" || fileType == "request_update" || fileType == "migration" {
		formatted, err := format.Source([]byte(content))
		if err == nil {
			return string(formatted), nil
		}
	}

	return content, nil
}

func (s *CodeGeneratorServiceImpl) Save(moduleName, tableName string, fields []FieldConfig, selectedFiles []string, options map[string]bool) ([]string, error) {
	files, err := s.Generate(moduleName, tableName, fields, selectedFiles, options)
	if err != nil {
		return nil, err
	}

	var existingFiles []string
	for _, file := range files {
		if _, err := os.Stat(file.Path); err == nil {
			existingFiles = append(existingFiles, file.Path)
		}
	}

	if len(existingFiles) > 0 {
		return nil, &FilesExistError{
			Files: existingFiles,
		}
	}

	var savedFiles []string
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := os.WriteFile(file.Path, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}

		savedFiles = append(savedFiles, file.Path)
	}

	return savedFiles, nil
}

func (s *CodeGeneratorServiceImpl) ForceSave(moduleName, tableName string, fields []FieldConfig, selectedFiles []string, options map[string]bool) ([]string, error) {
	files, err := s.Generate(moduleName, tableName, fields, selectedFiles, options)
	if err != nil {
		return nil, err
	}

	var savedFiles []string
	for _, file := range files {
		dir := filepath.Dir(file.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := os.WriteFile(file.Path, []byte(file.Content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", file.Path, err)
		}

		savedFiles = append(savedFiles, file.Path)
	}

	return savedFiles, nil
}

func (s *CodeGeneratorServiceImpl) GetFieldTypes() []FieldType {
	return []FieldType{
		{
			Label:      "字符串",
			Value:      "string",
			GoType:     "string",
			DBType:     "string",
			Validators: []string{"string", "max:255"},
		},
		{
			Label:      "文本",
			Value:      "text",
			GoType:     "string",
			DBType:     "text",
			Validators: []string{"string"},
		},
		{
			Label:      "整数",
			Value:      "integer",
			GoType:     "int",
			DBType:     "integer",
			Validators: []string{"integer"},
		},
		{
			Label:      "小数",
			Value:      "decimal",
			GoType:     "float64",
			DBType:     "decimal",
			Validators: []string{"numeric"},
		},
		{
			Label:      "布尔值",
			Value:      "boolean",
			GoType:     "bool",
			DBType:     "boolean",
			Validators: []string{"boolean"},
		},
		{
			Label:      "日期",
			Value:      "date",
			GoType:     "time.Time",
			DBType:     "date",
			Validators: []string{"date"},
		},
		{
			Label:      "日期时间",
			Value:      "datetime",
			GoType:     "time.Time",
			DBType:     "datetime",
			Validators: []string{"date"},
		},
		{
			Label:      "时间戳",
			Value:      "timestamp",
			GoType:     "int64",
			DBType:     "timestamp",
			Validators: []string{"integer"},
		},
		{
			Label:      "JSON",
			Value:      "json",
			GoType:     "string",
			DBType:     "json",
			Validators: []string{"json"},
		},
	}
}

func (s *CodeGeneratorServiceImpl) getTemplateName(fileType string) (string, error) {
	switch fileType {
	case "model":
		return "templates/model.tpl", nil
	case "controller":
		return "templates/controller.tpl", nil
	case "service":
		return "templates/service.tpl", nil
	case "request_create":
		return "templates/request_create.tpl", nil
	case "request_update":
		return "templates/request_update.tpl", nil
	case "migration":
		return "templates/migration.tpl", nil
	case "api":
		return "templates/api.js.tpl", nil
	case "list_page":
		return "templates/list_page.vue.tpl", nil
	case "form_page":
		return "templates/form_page.vue.tpl", nil
	default:
		return "", fmt.Errorf("unknown file type: %s", fileType)
	}
}

func (s *CodeGeneratorServiceImpl) buildTemplateData(moduleName, tableName string, fields []FieldConfig, fileType string, options map[string]bool) interface{} {
	templateFields := s.convertFieldsToTemplateFields(fields)

	// Default options
	hasCreate := true
	hasEdit := true
	hasDelete := true

	if options != nil {
		if val, ok := options["has_create"]; ok {
			hasCreate = val
		}
		if val, ok := options["has_edit"]; ok {
			hasEdit = val
		}
		if val, ok := options["has_delete"]; ok {
			hasDelete = val
		}
	}

	switch fileType {
	case "model":
		return struct {
			ModelName string
			TableName string
			Fields    []TemplateFieldConfig
		}{
			ModelName: toPascalCase(moduleName),
			TableName: tableName,
			Fields:    templateFields,
		}
	case "controller":
		var searchableFields []TemplateFieldConfig
		var listFields []TemplateFieldConfig
		for _, field := range templateFields {
			if field.Searchable {
				searchableFields = append(searchableFields, field)
			}
			if field.Relation == nil {
				listFields = append(listFields, field)
			}
		}
		return struct {
			ControllerName    string
			ServiceName       string
			ModelName         string
			ModuleName        string
			SearchableFields  []TemplateFieldConfig
			RequestCreateName string
			RequestUpdateName string
			HasCreate         bool
			HasEdit           bool
			HasDelete         bool
		}{
			ControllerName:    toPascalCase(moduleName) + "Controller",
			ServiceName:       toPascalCase(moduleName) + "Service",
			ModelName:         toPascalCase(moduleName),
			ModuleName:        moduleName,
			SearchableFields:  searchableFields,
			RequestCreateName: toPascalCase(moduleName) + "Create",
			RequestUpdateName: toPascalCase(moduleName) + "Update",
			HasCreate:         hasCreate,
			HasEdit:           hasEdit,
			HasDelete:         hasDelete,
		}
	case "service":
		var searchableFields []TemplateFieldConfig
		for _, field := range templateFields {
			if field.Searchable {
				searchableFields = append(searchableFields, field)
			}
		}
		return struct {
			ServiceName       string
			ModelName         string
			ModuleName        string
			SearchableFields  []TemplateFieldConfig
			RequestCreateName string
			RequestUpdateName string
			FormFields        []TemplateFieldConfig
			HasCreate         bool
			HasEdit           bool
			HasDelete         bool
		}{
			ServiceName:       toPascalCase(moduleName) + "Service",
			ModelName:         toPascalCase(moduleName),
			ModuleName:        moduleName,
			SearchableFields:  searchableFields,
			RequestCreateName: toPascalCase(moduleName) + "Create",
			RequestUpdateName: toPascalCase(moduleName) + "Update",
			FormFields:        templateFields,
			HasCreate:         hasCreate,
			HasEdit:           hasEdit,
			HasDelete:         hasDelete,
		}
	case "request_create":
		return struct {
			ModuleName        string
			TableName         string
			FormFields        []TemplateFieldConfig
			RequestCreateName string
		}{
			ModuleName:        moduleName,
			TableName:         tableName,
			FormFields:        templateFields,
			RequestCreateName: toPascalCase(moduleName) + "Create",
		}
	case "request_update":
		return struct {
			ModuleName        string
			TableName         string
			FormFields        []TemplateFieldConfig
			RequestUpdateName string
		}{
			ModuleName:        moduleName,
			TableName:         tableName,
			FormFields:        templateFields,
			RequestUpdateName: toPascalCase(moduleName) + "Update",
		}
	case "migration":
		for i := range templateFields {
			templateFields[i].MigrationMethod = getMigrationMethod(templateFields[i].DBType)
		}
		return struct {
			ModelName string
			TableName string
			Fields    []TemplateFieldConfig
		}{
			ModelName: toPascalCase(moduleName),
			TableName: tableName,
			Fields:    templateFields,
		}
	case "api":
		return struct {
			ModelName   string
			ModuleName  string
			ModuleNameK string
			FormFields  []TemplateFieldConfig
			HasCreate   bool
			HasEdit     bool
			HasDelete   bool
		}{
			ModelName:   toPascalCase(moduleName),
			ModuleName:  moduleName,
			ModuleNameK: toKebabCase(moduleName),
			FormFields:  templateFields,
			HasCreate:   hasCreate,
			HasEdit:     hasEdit,
			HasDelete:   hasDelete,
		}
	case "list_page":
		var searchableFields []TemplateFieldConfig
		var listFields []TemplateFieldConfig
		for _, field := range templateFields {
			if field.Searchable {
				searchableFields = append(searchableFields, field)
			}
			if field.Relation == nil {
				listFields = append(listFields, field)
			}
		}
		return struct {
			ModelName        string
			ModuleName       string
			ModuleNameK      string
			SearchableFields []TemplateFieldConfig
			ListFields       []TemplateFieldConfig
			FormFields       []TemplateFieldConfig
			HasCreate        bool
			HasEdit          bool
			HasDelete        bool
		}{
			ModelName:        toPascalCase(moduleName),
			ModuleName:       moduleName,
			ModuleNameK:      toKebabCase(moduleName),
			SearchableFields: searchableFields,
			ListFields:       listFields,
			FormFields:       templateFields,
			HasCreate:        hasCreate,
			HasEdit:          hasEdit,
			HasDelete:        hasDelete,
		}
	case "form_page":
		return struct {
			ModelName   string
			ModuleName  string
			ModuleNameK string
			FormFields  []TemplateFieldConfig
			HasCreate   bool
			HasEdit     bool
		}{
			ModelName:   toPascalCase(moduleName),
			ModuleName:  moduleName,
			ModuleNameK: toKebabCase(moduleName),
			FormFields:  templateFields,
			HasCreate:   hasCreate,
			HasEdit:     hasEdit,
		}
	default:
		return struct {
			ModuleName        string
			TableName         string
			FormFields        []TemplateFieldConfig
			RequestCreateName string
			RequestUpdateName string
		}{
			ModuleName:        moduleName,
			TableName:         tableName,
			FormFields:        templateFields,
			RequestCreateName: toPascalCase(moduleName) + "Create",
			RequestUpdateName: toPascalCase(moduleName) + "Update",
		}
	}
}

type TemplateFieldConfig struct {
	Name            string
	Label           string
	FieldName       string
	JsonName        string
	GoType          string
	DBType          string
	Validators      []string
	Required        bool
	MigrationMethod string
	PascalName      string
	Comment         string
	FormType        string
	SearchType      string
	SearchUIType    string
	Dictionary      string
	ApiUrl          string
	Sortable        bool
	Searchable      bool
	Relation        *RelationConfig
}

func (s *CodeGeneratorServiceImpl) convertFieldsToTemplateFields(fields []FieldConfig) []TemplateFieldConfig {
	fieldTypes := s.GetFieldTypes()
	fieldTypeMap := make(map[string]FieldType)
	for _, ft := range fieldTypes {
		fieldTypeMap[ft.Value] = ft
	}

	templateFields := make([]TemplateFieldConfig, len(fields))
	for i, field := range fields {
		fieldType, exists := fieldTypeMap[field.DBType]
		if !exists {
			fieldType = fieldTypes[0]
		}

		templateFields[i] = TemplateFieldConfig{
			Name:         field.Name,
			Label:        field.Label,
			FieldName:    toPascalCase(field.Name),
			JsonName:     field.Name,
			GoType:       fieldType.GoType,
			DBType:       field.DBType,
			Validators:   field.Validators,
			Required:     field.Required,
			PascalName:   toPascalCase(field.Name),
			Comment:      field.Label,
			FormType:     getFormType(field.DBType),
			SearchType:   getSearchType(field.SearchType),
			SearchUIType: getSearchUIType(field.SearchUIType, field.DBType),
			Dictionary:   field.Dictionary,
			ApiUrl:       getApiUrl(field.Dictionary, field.ApiUrl),
			Sortable:     getSortable(field.DBType),
			Searchable:   field.Searchable,
			Relation:     field.Relation,
		}
	}

	return templateFields
}

func (s *CodeGeneratorServiceImpl) generateModel(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/model.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read model template: %w", err)
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ModelName string
		TableName string
		Fields    []TemplateFieldConfig
	}{
		ModelName: toPascalCase(moduleName),
		TableName: tableName,
		Fields:    templateFields,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("app/models/%s.go", toSnakeCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateController(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/controller.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read controller template: %w", err)
	}

	// Default options
	hasCreate := true
	hasEdit := true
	hasDelete := true

	if options != nil {
		if val, ok := options["has_create"]; ok {
			hasCreate = val
		}
		if val, ok := options["has_edit"]; ok {
			hasEdit = val
		}
		if val, ok := options["has_delete"]; ok {
			hasDelete = val
		}
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ControllerName    string
		ServiceName       string
		ModelName         string
		ModuleName        string
		SearchableFields  []TemplateFieldConfig
		RequestCreateName string
		RequestUpdateName string
		HasCreate         bool
		HasEdit           bool
		HasDelete         bool
	}{
		ControllerName:    toPascalCase(moduleName) + "Controller",
		ServiceName:       toPascalCase(moduleName) + "Service",
		ModelName:         toPascalCase(moduleName),
		ModuleName:        moduleName,
		SearchableFields:  templateFields,
		RequestCreateName: toPascalCase(moduleName) + "Create",
		RequestUpdateName: toPascalCase(moduleName) + "Update",
		HasCreate:         hasCreate,
		HasEdit:           hasEdit,
		HasDelete:         hasDelete,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("app/http/controllers/admin/%s_controller.go", toSnakeCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateService(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/service.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read service template: %w", err)
	}

	// Default options
	hasCreate := true
	hasEdit := true
	hasDelete := true

	if options != nil {
		if val, ok := options["has_create"]; ok {
			hasCreate = val
		}
		if val, ok := options["has_edit"]; ok {
			hasEdit = val
		}
		if val, ok := options["has_delete"]; ok {
			hasDelete = val
		}
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ServiceName       string
		ModelName         string
		ModuleName        string
		SearchableFields  []TemplateFieldConfig
		RequestCreateName string
		RequestUpdateName string
		FormFields        []TemplateFieldConfig
		HasCreate         bool
		HasEdit           bool
		HasDelete         bool
	}{
		ServiceName:       toPascalCase(moduleName) + "Service",
		ModelName:         toPascalCase(moduleName),
		ModuleName:        moduleName,
		SearchableFields:  templateFields,
		RequestCreateName: toPascalCase(moduleName) + "Create",
		RequestUpdateName: toPascalCase(moduleName) + "Update",
		FormFields:        templateFields,
		HasCreate:         hasCreate,
		HasEdit:           hasEdit,
		HasDelete:         hasDelete,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("app/services/%s_service.go", toSnakeCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateRequestCreate(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/request_create.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read request_create template: %w", err)
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ModuleName        string
		TableName         string
		FormFields        []TemplateFieldConfig
		RequestCreateName string
	}{
		ModuleName:        moduleName,
		TableName:         tableName,
		FormFields:        templateFields,
		RequestCreateName: toPascalCase(moduleName) + "Create",
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("app/http/requests/admin/%s_create.go", toSnakeCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateRequestUpdate(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/request_update.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read request_update template: %w", err)
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ModuleName        string
		TableName         string
		FormFields        []TemplateFieldConfig
		RequestUpdateName string
	}{
		ModuleName:        moduleName,
		TableName:         tableName,
		FormFields:        templateFields,
		RequestUpdateName: toPascalCase(moduleName) + "Update",
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("app/http/requests/admin/%s_update.go", toSnakeCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateMigration(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/migration.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read migration template: %w", err)
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	for i := range templateFields {
		templateFields[i].MigrationMethod = getMigrationMethod(templateFields[i].DBType)
	}

	data := struct {
		ModelName string
		TableName string
		Fields    []TemplateFieldConfig
	}{
		ModelName: toPascalCase(moduleName),
		TableName: tableName,
		Fields:    templateFields,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("database/migrations/create_%s_table.go", tableName),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateFrontendAPI(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/api.js.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read api template: %w", err)
	}

	// Default options
	hasCreate := true
	hasEdit := true
	hasDelete := true

	if options != nil {
		if val, ok := options["has_create"]; ok {
			hasCreate = val
		}
		if val, ok := options["has_edit"]; ok {
			hasEdit = val
		}
		if val, ok := options["has_delete"]; ok {
			hasDelete = val
		}
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ModelName   string
		ModuleName  string
		ModuleNameK string
		FormFields  []TemplateFieldConfig
		HasCreate   bool
		HasEdit     bool
		HasDelete   bool
	}{
		ModelName:   toPascalCase(moduleName),
		ModuleName:  moduleName,
		ModuleNameK: toKebabCase(moduleName),
		FormFields:  templateFields,
		HasCreate:   hasCreate,
		HasEdit:     hasEdit,
		HasDelete:   hasDelete,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("html/src/api/%s.js", toKebabCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateFrontendListPage(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/list_page.vue.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read list_page template: %w", err)
	}

	// Default options
	hasCreate := true
	hasEdit := true
	hasDelete := true

	if options != nil {
		if val, ok := options["has_create"]; ok {
			hasCreate = val
		}
		if val, ok := options["has_edit"]; ok {
			hasEdit = val
		}
		if val, ok := options["has_delete"]; ok {
			hasDelete = val
		}
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ModelName        string
		ModuleName       string
		ModuleNameK      string
		SearchableFields []TemplateFieldConfig
		ListFields       []TemplateFieldConfig
		FormFields       []TemplateFieldConfig
		HasCreate        bool
		HasEdit          bool
		HasDelete        bool
	}{
		ModelName:        toPascalCase(moduleName),
		ModuleName:       moduleName,
		ModuleNameK:      toKebabCase(moduleName),
		SearchableFields: templateFields,
		ListFields:       templateFields,
		FormFields:       templateFields,
		HasCreate:        hasCreate,
		HasEdit:          hasEdit,
		HasDelete:        hasDelete,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("html/src/views/%s/%sList.vue", toKebabCase(moduleName), toPascalCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) generateFrontendFormPage(moduleName, tableName string, fields []FieldConfig, options map[string]bool) (GeneratedFile, error) {
	templateContent, err := templates.ReadFile("templates/form_page.vue.tpl")
	if err != nil {
		return GeneratedFile{}, fmt.Errorf("failed to read form_page template: %w", err)
	}

	// Default options
	hasCreate := true
	hasEdit := true

	if options != nil {
		if val, ok := options["has_create"]; ok {
			hasCreate = val
		}
		if val, ok := options["has_edit"]; ok {
			hasEdit = val
		}
	}

	templateFields := s.convertFieldsToTemplateFields(fields)
	data := struct {
		ModelName   string
		ModuleName  string
		ModuleNameK string
		FormFields  []TemplateFieldConfig
		HasCreate   bool
		HasEdit     bool
	}{
		ModelName:   toPascalCase(moduleName),
		ModuleName:  moduleName,
		ModuleNameK: toKebabCase(moduleName),
		FormFields:  templateFields,
		HasCreate:   hasCreate,
		HasEdit:     hasEdit,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("html/src/views/%s/%sForm.vue", toKebabCase(moduleName), toPascalCase(moduleName)),
		Content: content,
	}, nil
}

func (s *CodeGeneratorServiceImpl) executeTemplate(templateContent string, data interface{}) (string, error) {
	tmpl, err := template.New("code").Delims("<<", ">>").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return builder.String(), nil
}

func toSnakeCase(s string) string {
	result := ""
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += string(r)
	}
	return strings.ToLower(result)
}

func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i := range words {
		if len(words[i]) > 0 {
			words[i] = strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
		}
	}
	return strings.Join(words, "")
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}

func getMigrationMethod(dbType string) string {
	switch dbType {
	case "string":
		return "String"
	case "text":
		return "Text"
	case "integer":
		return "Integer"
	case "decimal":
		return "Decimal"
	case "boolean":
		return "Boolean"
	case "date":
		return "Date"
	case "datetime":
		return "DateTime"
	case "timestamp":
		return "Timestamp"
	case "json":
		return "Json"
	default:
		return "String"
	}
}

func getFormType(dbType string) string {
	switch dbType {
	case "string":
		return "input"
	case "text":
		return "textarea"
	case "integer":
		return "input"
	case "decimal":
		return "input"
	case "boolean":
		return "switch"
	case "date":
		return "date-picker"
	case "datetime":
		return "datetime-picker"
	case "timestamp":
		return "datetime-picker"
	case "json":
		return "textarea"
	default:
		return "input"
	}
}

func getSearchType(searchType string) string {
	if searchType == "" {
		return "like"
	}
	return searchType
}

func getSearchUIType(searchUIType string, dbType string) string {
	if searchUIType != "" {
		return searchUIType
	}
	switch dbType {
	case "date":
		return "date"
	case "datetime", "timestamp":
		return "datetime"
	case "boolean":
		return "select"
	default:
		return "input"
	}
}

func getApiUrl(dictionary string, apiUrl string) string {
	if apiUrl != "" {
		return apiUrl
	}
	if dictionary != "" {
		return "/options?type=" + dictionary
	}
	return ""
}

func getSortable(dbType string) bool {
	switch dbType {
	case "string", "integer", "decimal", "date", "datetime", "timestamp":
		return true
	case "text", "boolean", "json":
		return false
	default:
		return true
	}
}
