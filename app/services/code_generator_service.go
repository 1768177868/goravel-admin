package services

import (
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/goravel/framework/facades"
	"gorm.io/gorm"
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
	ShowInList   bool            `json:"show_in_list"`
	ShowInForm   bool            `json:"show_in_form"`
	ShowInDetail bool            `json:"show_in_detail"`
	Precision    int             `json:"precision"` // 精度 (总位数)
	Scale        int             `json:"scale"`     // 标度 (小数位数)
	FormType     string          `json:"form_type"`
}

type RelationConfig struct {
	Table        string `json:"table"`         // 关联表名
	ForeignKey   string `json:"foreign_key"`   // 外键字段
	DisplayField string `json:"display_field"` // 显示字段
	RelationType string `json:"relation_type"` // 关联类型：hasOne, belongsTo, hasMany
	Alias        string `json:"alias"`         // 别名
	IsTree       bool   `json:"is_tree"`       // 是否为树形数据
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
	GetTables() ([]string, error)
	GetTableColumns(tableName string) ([]FieldConfig, error)
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

func (s *CodeGeneratorServiceImpl) getGormDB() (*gorm.DB, error) {
	orm := facades.Orm()
	if orm == nil {
		return nil, fmt.Errorf("ORM facade is nil")
	}

	// Try to get Query()
	query := orm.Query()
	if query == nil {
		return nil, fmt.Errorf("Query() returned nil")
	}

	// Use reflection to call Instance() on Query object
	// Because Query interface in Goravel framework might not expose Instance() method directly
	// but the implementation (gorm.Query) does have it.
	queryValue := reflect.ValueOf(query)
	instanceMethod := queryValue.MethodByName("Instance")
	if !instanceMethod.IsValid() {
		// Try to see if it's wrapped
		if queryValue.Kind() == reflect.Interface {
			queryValue = queryValue.Elem()
			instanceMethod = queryValue.MethodByName("Instance")
		}
	}

	if instanceMethod.IsValid() {
		results := instanceMethod.Call(nil)
		if len(results) > 0 {
			if db, ok := results[0].Interface().(*gorm.DB); ok {
				return db, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to get *gorm.DB from ORM")
}

func (s *CodeGeneratorServiceImpl) GetTables() ([]string, error) {
	db, err := s.getGormDB()
	if err != nil {
		return nil, err
	}

	var tables []string
	if err := db.Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		return nil, err
	}

	// 过滤掉系统默认表
	ignoreTables := map[string]bool{
		"users":                  true,
		"roles":                  true,
		"permissions":            true,
		"menus":                  true,
		"role_users":             true,
		"role_permissions":       true,
		"admin_role":             true,
		"role_menu":              true,
		"role_permission":        true,
		"migrations":             true,
		"personal_access_tokens": true,
		"settings":               true,
		"dict_types":             true,
		"dict_data":              true,
		"admins":                 true,
		"failed_jobs":            true,
		"jobs":                   true,
		"attachments":            true,
		"blacklists":             true,
		"configs":                true,
		"currencies":             true,
		"departments":            true,
		"dictionaries":           true,
		"exports":                true,
		"login_logs":             true,
		"notifications":          true,
		"operation_logs":         true,
		"system_logs":            true,
		"payment_methods":        true,
		"payments":               true,
	}

	var filteredTables []string
	for _, table := range tables {
		if ignoreTables[table] {
			continue
		}
		// 过滤分表逻辑：
		// 1. 过滤掉包含年份月份后缀的表 (例如 _202512, _202601)
		// 2. 过滤掉 Hash 分表 (例如 _1, _2, _100)
		isShardedTable := false

		// 查找最后一个下划线
		lastUnderscoreIndex := strings.LastIndex(table, "_")
		if lastUnderscoreIndex != -1 && lastUnderscoreIndex < len(table)-1 {
			suffix := table[lastUnderscoreIndex+1:]

			// 检查后缀是否纯数字
			isNumeric := true
			for _, ch := range suffix {
				if ch < '0' || ch > '9' {
					isNumeric = false
					break
				}
			}

			if isNumeric {
				isShardedTable = true
			}
		}

		if !isShardedTable {
			filteredTables = append(filteredTables, table)
		}
	}

	return filteredTables, nil
}

func (s *CodeGeneratorServiceImpl) GetTableColumns(tableName string) ([]FieldConfig, error) {
	db, err := s.getGormDB()
	if err != nil {
		return nil, err
	}

	columnTypes, err := db.Migrator().ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}

	var fields []FieldConfig
	for _, ct := range columnTypes {
		name := ct.Name()
		if name == "deleted_at" || name == "id" {
			continue
		}
		dbType := ct.DatabaseTypeName()

		// Skip internal fields if needed, but we usually want them
		// Mapping logic
		goType := "string"
		jsonDBType := "string"

		lowerDBType := strings.ToLower(dbType)
		if strings.Contains(lowerDBType, "int") {
			if strings.Contains(lowerDBType, "bigint") {
				goType = "int64"
				jsonDBType = "bigInteger"
			} else if strings.Contains(lowerDBType, "tinyint") {
				goType = "int"
				jsonDBType = "unsignedTinyInteger"
			} else {
				goType = "int"
				jsonDBType = "integer"
			}
		} else if strings.Contains(lowerDBType, "char") {
			goType = "string"
			jsonDBType = "string"
		} else if strings.Contains(lowerDBType, "text") {
			goType = "string"
			jsonDBType = "text"
		} else if strings.Contains(lowerDBType, "datetime") || strings.Contains(lowerDBType, "timestamp") {
			goType = "time.Time"
			jsonDBType = "datetime"
		} else if strings.Contains(lowerDBType, "date") {
			goType = "time.Time"
			jsonDBType = "date"
		} else if strings.Contains(lowerDBType, "decimal") || strings.Contains(lowerDBType, "double") || strings.Contains(lowerDBType, "float") {
			goType = "float64"
			jsonDBType = "decimal"
		} else if strings.Contains(lowerDBType, "bool") {
			goType = "bool"
			jsonDBType = "boolean"
		} else if strings.Contains(lowerDBType, "json") {
			goType = "string"
			jsonDBType = "json"
		}

		// Refine based on column name
		if name == "id" || name == "created_at" || name == "updated_at" {
			jsonDBType = "datetime"
			if name == "id" {
				jsonDBType = "bigInteger"
				goType = "int64"
			} else {
				goType = "time.Time"
			}
		}

		// Skip internal fields if needed, but we usually want them
		if name == "deleted_at" || name == "id" {
			continue
		}

		// SearchType default
		searchType := "="
		if goType == "string" {
			searchType = "like"
		}

		// Label defaults to name
		label := name
		comment, _ := ct.Comment()
		if comment != "" {
			label = comment
		}

		nullable, _ := ct.Nullable()

		// Precision/Scale
		precision, scale, _ := ct.DecimalSize()

		// 系统字段默认不在表单中显示
		showInForm := true
		if name == "created_at" || name == "updated_at" || name == "deleted_at" {
			showInForm = false
		}

		fields = append(fields, FieldConfig{
			Name:         name,
			Label:        label,
			GoType:       goType,
			DBType:       jsonDBType, // 这里是后端推断的类型，前端会根据这个去匹配 fieldTypes
			Required:     !nullable,
			Searchable:   true,
			Sortable:     false,
			ShowInList:   true,
			ShowInForm:   showInForm,
			ShowInDetail: true,
			SearchType:   searchType,
			SearchUIType: getSearchUIType("", jsonDBType),
			Dictionary:   "",
			ApiUrl:       "",
			Precision:    int(precision),
			Scale:        int(scale),
		})

		// 智能推断 FormType 和 Relation
		field := &fields[len(fields)-1]

		// 1. 推断 FormType
		if field.Name == "id" || field.Name == "created_at" || field.Name == "updated_at" || field.Name == "deleted_at" {
			// 系统字段，通常不需要在表单中显示，或者只读
			field.FormType = "input"
		} else if strings.HasSuffix(field.Name, "_id") {
			field.FormType = "select"
			// 推断关联
			refTable := strings.TrimSuffix(field.Name, "_id") + "s" // 简单复数化
			// 树形数据只有在自定义API时才可能，默认关联不使用树形
			field.Relation = &RelationConfig{
				Table:        refTable,
				ForeignKey:   field.Name,
				DisplayField: "name", // 默认猜测 name
				RelationType: "belongsTo",
				Alias:        "",    // 默认使用表名转 PascalCase
				IsTree:       false, // 默认不是树形，只有自定义API时才可能是树形
			}
		} else if strings.Contains(field.Name, "image") || strings.Contains(field.Name, "avatar") || strings.Contains(field.Name, "photo") || strings.Contains(field.Name, "pic") {
			field.FormType = "image-upload"
		} else if strings.Contains(field.Name, "file") {
			field.FormType = "file-upload"
		} else if (strings.Contains(field.Name, "content") || strings.Contains(field.Name, "description") || strings.Contains(field.Name, "detail")) && field.DBType == "text" {
			field.FormType = "editor"
		} else {
			field.FormType = getFormType(field.DBType)
		}
	}
	return fields, nil
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
			Label:      "大整数",
			Value:      "bigInteger",
			GoType:     "int64",
			DBType:     "bigInteger",
			Validators: []string{"integer"},
		},
		{
			Label:      "无符号大整数",
			Value:      "unsignedBigInteger",
			GoType:     "uint64",
			DBType:     "unsignedBigInteger",
			Validators: []string{"integer", "min:0"},
		},
		{
			Label:      "无符号小整数",
			Value:      "unsignedTinyInteger",
			GoType:     "uint8",
			DBType:     "unsignedTinyInteger",
			Validators: []string{"integer", "min:0", "max:255"},
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

func (s *CodeGeneratorServiceImpl) buildTemplateData(moduleName, tableName string, fields []FieldConfig, fileType string, options map[string]bool) any {
	templateFields := s.convertFieldsToTemplateFields(fields)

	// Default options
	hasCreate := true
	hasEdit := true
	hasDelete := true
	hasExport := false

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
		if val, ok := options["has_export"]; ok {
			hasExport = val
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
			HasExport   bool
		}{
			ModelName:   toPascalCase(moduleName),
			ModuleName:  moduleName,
			ModuleNameK: toKebabCase(moduleName),
			FormFields:  templateFields,
			HasCreate:   hasCreate,
			HasEdit:     hasEdit,
			HasDelete:   hasDelete,
			HasExport:   hasExport,
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
			HasExport        bool
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
			HasExport:        hasExport,
		}
	case "form_page":
		// 检查是否有 editor 类型的字段
		hasEditor := false
		for _, field := range templateFields {
			if field.FormType == "editor" && field.ShowInForm {
				hasEditor = true
				break
			}
		}

		// 检查是否有 markdown 类型的字段
		hasMarkdown := false
		for _, field := range templateFields {
			if field.FormType == "markdown" && field.ShowInForm {
				hasMarkdown = true
				break
			}
		}

		// 检查是否有 image-upload 类型的字段
		hasImageUpload := false
		for _, field := range templateFields {
			if field.FormType == "image-upload" && field.ShowInForm {
				hasImageUpload = true
				break
			}
		}

		return struct {
			ModelName      string
			ModuleName     string
			ModuleNameK    string
			FormFields     []TemplateFieldConfig
			HasCreate      bool
			HasEdit        bool
			HasEditor      bool
			HasMarkdown    bool
			HasImageUpload bool
		}{
			ModelName:      toPascalCase(moduleName),
			ModuleName:     moduleName,
			ModuleNameK:    toKebabCase(moduleName),
			FormFields:     templateFields,
			HasCreate:      hasCreate,
			HasEdit:        hasEdit,
			HasEditor:      hasEditor,
			HasMarkdown:    hasMarkdown,
			HasImageUpload: hasImageUpload,
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
	ShowInList      bool
	ShowInForm      bool
	ShowInDetail    bool
	Relation        *TemplateRelationConfig
	IsTree          bool // 是否为树形数据（用于自定义API）
	Precision       int
	Scale           int
}

type TemplateRelationConfig struct {
	*RelationConfig
	Name      string // PascalCase 关联名 (如: Admin)
	JsonName  string // camelCase 关联名 (如: admin)
	ModelName string // 关联表对应的模型名称 (如: Admin)
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

		var relation *TemplateRelationConfig
		if field.Relation != nil {
			var relationName string

			// 1. 如果有别名，直接使用别名
			if field.Relation.Alias != "" {
				relationName = toPascalCase(field.Relation.Alias)
			} else {
				// 2. 否则优先尝试从字段名推导 (如 AdminID -> Admin)
				relationName = toPascalCase(field.Name)
				if strings.HasSuffix(relationName, "ID") {
					relationName = strings.TrimSuffix(relationName, "ID")
				} else if strings.HasSuffix(relationName, "Id") {
					relationName = strings.TrimSuffix(relationName, "Id")
				} else {
					// 3. 最后使用表名
					relationName = toPascalCase(field.Relation.Table)
					// 如果是 belongsTo 或 hasOne，尝试转为单数
					if (field.Relation.RelationType == "belongsTo" || field.Relation.RelationType == "hasOne") && strings.HasSuffix(relationName, "s") {
						relationName = relationName[:len(relationName)-1]
					}
				}
			}

			// 计算模型名称 (始终基于表名)
			modelName := toPascalCase(field.Relation.Table)
			// 如果是 belongsTo 或 hasOne，尝试转为单数
			if (field.Relation.RelationType == "belongsTo" || field.Relation.RelationType == "hasOne") && strings.HasSuffix(modelName, "s") {
				modelName = modelName[:len(modelName)-1]
			}

			// 计算 JSON 名称 (首字母小写)
			jsonName := strings.ToLower(relationName[:1]) + relationName[1:]

			relation = &TemplateRelationConfig{
				RelationConfig: field.Relation,
				Name:           relationName,
				JsonName:       jsonName,
				ModelName:      modelName,
			}
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
			Comment:      field.Label, // Use Label as Comment
			FormType:     field.FormType,
			SearchType:   getSearchType(field.SearchType),
			SearchUIType: getSearchUIType(field.SearchUIType, field.DBType),
			Dictionary:   field.Dictionary,
			ApiUrl:       getApiUrl(field.Dictionary, field.ApiUrl),
			Sortable:     field.Sortable,
			Searchable:   field.Searchable,
			ShowInList:   field.ShowInList,
			ShowInForm:   field.ShowInForm,
			ShowInDetail: field.ShowInDetail,
			Relation:     relation,
			IsTree:       field.Relation != nil && field.Relation.IsTree, // 从 relation 中读取 is_tree
			Precision:    field.Precision,
			Scale:        field.Scale,
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

	timestamp := time.Now().Format("20060102150405")
	data := struct {
		ModelName string
		TableName string
		Fields    []TemplateFieldConfig
		Timestamp string
	}{
		ModelName: toPascalCase(moduleName),
		TableName: tableName,
		Fields:    templateFields,
		Timestamp: timestamp,
	}

	content, err := s.executeTemplate(string(templateContent), data)
	if err != nil {
		return GeneratedFile{}, err
	}

	return GeneratedFile{
		Path:    fmt.Sprintf("database/migrations/%s_create_%s_table.go", timestamp, toSnakeCase(tableName)),
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
	hasExport := false

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
		if val, ok := options["has_export"]; ok {
			hasExport = val
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
		HasExport   bool
	}{
		ModelName:   toPascalCase(moduleName),
		ModuleName:  moduleName,
		ModuleNameK: toKebabCase(moduleName),
		FormFields:  templateFields,
		HasCreate:   hasCreate,
		HasEdit:     hasEdit,
		HasDelete:   hasDelete,
		HasExport:   hasExport,
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
	hasExport := false

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
		if val, ok := options["has_export"]; ok {
			hasExport = val
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
		HasExport        bool
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
		HasExport:        hasExport,
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

	// 检查是否有 editor 类型的字段
	hasEditor := false
	for _, field := range templateFields {
		if field.FormType == "editor" && field.ShowInForm {
			hasEditor = true
			break
		}
	}

	// 检查是否有 markdown 类型的字段
	hasMarkdown := false
	for _, field := range templateFields {
		if field.FormType == "markdown" && field.ShowInForm {
			hasMarkdown = true
			break
		}
	}

	// 检查是否有 image-upload 类型的字段
	hasImageUpload := false
	for _, field := range templateFields {
		if field.FormType == "image-upload" && field.ShowInForm {
			hasImageUpload = true
			break
		}
	}

	data := struct {
		ModelName      string
		ModuleName     string
		ModuleNameK    string
		FormFields     []TemplateFieldConfig
		HasCreate      bool
		HasEdit        bool
		HasEditor      bool
		HasMarkdown    bool
		HasImageUpload bool
	}{
		ModelName:      toPascalCase(moduleName),
		ModuleName:     moduleName,
		ModuleNameK:    toKebabCase(moduleName),
		FormFields:     templateFields,
		HasCreate:      hasCreate,
		HasEdit:        hasEdit,
		HasEditor:      hasEditor,
		HasMarkdown:    hasMarkdown,
		HasImageUpload: hasImageUpload,
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

func (s *CodeGeneratorServiceImpl) executeTemplate(templateContent string, data any) (string, error) {
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
	case "bigInteger":
		return "BigInteger"
	case "unsignedBigInteger":
		return "UnsignedBigInteger"
	case "unsignedTinyInteger":
		return "UnsignedTinyInteger"
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
		return "/options?type=dictionary&dictionary_type=" + dictionary
	}
	return ""
}
