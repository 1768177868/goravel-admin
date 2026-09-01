package services

import "testing"

func TestBuildModuleManifest_CRUDPermissions(t *testing.T) {
	manifest, err := BuildModuleManifest("article", "articles", map[string]bool{
		"has_create": true,
		"has_edit":   true,
		"has_delete": true,
		"has_export": true,
	}, &ModuleInstallConfig{
		MenuTitle:       "文章管理",
		ParentMenuSlug:  "",
		MenuSort:        0,
	})
	if err != nil {
		t.Fatalf("BuildModuleManifest() error = %v", err)
	}

	if manifest.MenuSlug != "article" {
		t.Fatalf("MenuSlug = %q, want article", manifest.MenuSlug)
	}
	if manifest.Component != "article/ArticleList" {
		t.Fatalf("Component = %q, want article/ArticleList", manifest.Component)
	}
	if len(manifest.Permissions) != 6 {
		t.Fatalf("len(Permissions) = %d, want 6", len(manifest.Permissions))
	}
	if manifest.Permissions[0].Slug != "article.index" {
		t.Fatalf("first permission slug = %q", manifest.Permissions[0].Slug)
	}
}
