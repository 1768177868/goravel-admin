package services

import (
	"context"
	"strings"
	"testing"
)

func TestReactListPageTemplatePreview(t *testing.T) {
	s := NewCodeGeneratorService(context.Background())
	fields := []FieldConfig{
		{
			Name: "admin_id", Label: "Admin", DBType: "bigInteger", GoType: "int64",
			ShowInList: true, ShowInForm: true, FormType: "select",
			Relation: &RelationConfig{
				Table: "admins", ForeignKey: "admin_id", DisplayField: "name", RelationType: "belongsTo",
			},
		},
		{Name: "title", Label: "Title", ShowInList: true, ShowInForm: true, FormType: "input"},
		{Name: "content", Label: "Content", ShowInList: true, ShowInForm: true, FormType: "editor"},
		{Name: "status", Label: "Status", ShowInList: true, ShowInForm: true, FormType: "input"},
	}

	code, err := s.Preview("article", "articles", fields, "react_list_page", map[string]bool{
		"has_create": true, "has_edit": true, "has_delete": true, "show_toolbar": true,
	})
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	if !strings.Contains(code, "useListPage<ArticleRow>(") {
		t.Fatalf("expected typed useListPage call, got:\n%s", code[:min(500, len(code))])
	}
	if strings.Contains(code, "<<") || strings.Contains(code, ">>>") {
		t.Fatalf("template delimiters leaked into output")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
