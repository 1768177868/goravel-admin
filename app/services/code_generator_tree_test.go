package services

import (
	"strings"
	"testing"
)

func TestGenerateTreeModuleBackend(t *testing.T) {
	s := &CodeGeneratorServiceImpl{}
	fields := []FieldConfig{
		{Name: "name", GoType: "string", Label: "名称", ShowInList: true, ShowInForm: true, FormType: "input"},
		{Name: "parent_id", GoType: "uint", Label: "父级", ShowInList: false, ShowInForm: true, FormType: "number"},
		{Name: "sort", GoType: "int", Label: "排序", ShowInList: true, ShowInForm: true, FormType: "number"},
		{Name: "status", GoType: "uint8", Label: "状态", ShowInList: true, ShowInForm: true, FormType: "switch"},
	}
	options := map[string]bool{"is_tree_list": true}

	controller, err := s.generateController("category", "categories", fields, options)
	if err != nil {
		t.Fatalf("generate controller: %v", err)
	}
	if !strings.Contains(controller.Content, "GetTree()") {
		t.Fatal("controller should call GetTree for tree list")
	}
	if !strings.Contains(controller.Content, "hasCategoryTreeSearchFilters") {
		t.Fatal("controller should include tree search filter helper")
	}

	service, err := s.generateService("category", "categories", fields, options)
	if err != nil {
		t.Fatalf("generate service: %v", err)
	}
	if !strings.Contains(service.Content, "func (s *CategoryServiceImpl) GetTree()") {
		t.Fatal("service should include GetTree method")
	}

	model, err := s.generateModel("category", "categories", fields, options)
	if err != nil {
		t.Fatalf("generate model: %v", err)
	}
	if !strings.Contains(model.Content, "Children []Category") {
		t.Fatal("model should include Children field for tree list")
	}
}

func TestGenerateTreeModuleReactFrontend(t *testing.T) {
	s := &CodeGeneratorServiceImpl{}
	fields := []FieldConfig{
		{Name: "name", GoType: "string", Label: "名称", ShowInList: true, ShowInForm: true, FormType: "input"},
		{Name: "parent_id", GoType: "uint", Label: "父级", ShowInList: false, ShowInForm: true, FormType: "number"},
	}
	options := map[string]bool{"is_tree_list": true}

	listPage, err := s.generateReactListPage("category", "categories", fields, options)
	if err != nil {
		t.Fatalf("generate react list page: %v", err)
	}
	if !strings.Contains(listPage.Content, "normalizeCategoryTreeList") {
		t.Fatal("react tree list page should normalize tree data")
	}
	if !strings.Contains(listPage.Content, "expandedRowKeys") {
		t.Fatal("react tree list page should use expandable table")
	}
	if !strings.Contains(listPage.Content, "<Table<CategoryRow>") {
		t.Fatal("react tree list page Table generic should not have double >>")
	}

	config, err := s.generateReactListPageConfig("category", "categories", fields, options)
	if err != nil {
		t.Fatalf("generate react config: %v", err)
	}
	if !strings.Contains(config.Content, "buildCategoryTree") {
		t.Fatal("react tree config should build tree from flat list")
	}

	formModal, err := s.generateReactFormModal("category", "categories", fields, options)
	if err != nil {
		t.Fatalf("generate react form modal: %v", err)
	}
	if !strings.Contains(formModal.Content, "TreeSelect") {
		t.Fatal("react tree form modal should use TreeSelect for parent_id")
	}
}
