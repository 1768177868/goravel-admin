package services

import (
	"strings"
	"testing"
)

func TestGetReactListPageTemplateName_TreeList(t *testing.T) {
	s := &CodeGeneratorServiceImpl{}
	fields := []FieldConfig{{Name: "name", ShowInList: true, ShowInForm: true}}
	options := map[string]bool{"is_tree_list": true}

	got := s.getReactListPageTemplateName(fields, options)
	if got != "templates/react_tree_list_page.tsx.tpl" {
		t.Fatalf("template = %q, want react_tree_list_page", got)
	}

	config := s.getReactListPageConfigTemplateName(options)
	if config != "templates/react_tree_list.config.ts.tpl" {
		t.Fatalf("config template = %q", config)
	}

	form := s.getReactFormModalTemplateName(options)
	if form != "templates/react_tree_form_modal.tsx.tpl" {
		t.Fatalf("form template = %q", form)
	}
}

func TestResolveTreeLabelField(t *testing.T) {
	fields := []TemplateFieldConfig{
		{Name: "parent_id", ShowInList: true},
		{Name: "title", ShowInList: true},
	}
	if got := resolveTreeLabelField(fields); got != "title" {
		t.Fatalf("resolveTreeLabelField() = %q, want title", got)
	}
}

func TestGenerateReactListPage_TableGeneric(t *testing.T) {
	s := &CodeGeneratorServiceImpl{}
	fields := []FieldConfig{
		{Name: "name", GoType: "string", Label: "名称", ShowInList: true, ShowInForm: true, FormType: "input"},
		{Name: "status", GoType: "uint8", Label: "状态", ShowInList: true, ShowInForm: true, FormType: "select"},
	}
	listPage, err := s.generateReactListPage("article", "articles", fields, nil)
	if err != nil {
		t.Fatalf("generate react list page: %v", err)
	}
	if strings.Contains(listPage.Content, "<Table<ArticleRow>>") {
		t.Fatal("Table opening tag must not end with >>")
	}
	if !strings.Contains(listPage.Content, "<Table<ArticleRow>") {
		t.Fatal("Table opening tag missing")
	}
}
func TestGenerateReactFormModal_EditorField(t *testing.T) {
	s := &CodeGeneratorServiceImpl{}
	fields := []FieldConfig{
		{Name: "title", GoType: "string", Label: "标题", ShowInForm: true, FormType: "input"},
		{Name: "content", GoType: "string", Label: "内容", ShowInForm: true, FormType: "editor"},
	}
	formModal, err := s.generateReactFormModal("article", "articles", fields, map[string]bool{"create": true, "edit": true})
	if err != nil {
		t.Fatalf("generate react form modal: %v", err)
	}
	if !strings.Contains(formModal.Content, "WangEditor") {
		t.Fatal("editor field must render WangEditor")
	}
	if !strings.Contains(formModal.Content, "width={800}") {
		t.Fatal("editor modal should use wider layout")
	}
}

func TestGenerateReactListPage_EditorColumn(t *testing.T) {
	s := &CodeGeneratorServiceImpl{}
	fields := []FieldConfig{
		{Name: "title", GoType: "string", Label: "标题", ShowInList: true, ShowInForm: true, FormType: "input"},
		{Name: "content", GoType: "string", Label: "内容", ShowInList: true, ShowInForm: true, FormType: "editor"},
	}
	listPage, err := s.generateReactListPage("article", "articles", fields, nil)
	if err != nil {
		t.Fatalf("generate react list page: %v", err)
	}
	if !strings.Contains(listPage.Content, "extractTextFromMarkdown") {
		t.Fatal("editor list column must strip HTML for display")
	}
}

func TestApplyTreeListFieldFlags_ParentID(t *testing.T) {
	fields := []TemplateFieldConfig{{Name: "parent_id", FormType: "number", ShowInForm: true}}
	got := applyTreeListFieldFlags(fields, true)
	if !got[0].IsTree {
		t.Fatal("expected parent_id IsTree=true")
	}
	if got[0].FormType != "select" {
		t.Fatalf("FormType = %q, want select", got[0].FormType)
	}
}
