package utils

import (
	"goravel/app/models"
)

// DepartmentTreeItem 前端使用的部门树形结构
type DepartmentTreeItem struct {
	ID        uint                 `json:"id"`
	ParentID  uint                 `json:"parent_id"`
	Name      string               `json:"name"`
	Code      string               `json:"code,omitempty"`
	Leader    string               `json:"leader,omitempty"`
	Phone     string               `json:"phone,omitempty"`
	Email     string               `json:"email,omitempty"`
	Status    uint8                `json:"status"`
	Sort      int                  `json:"sort"`
	Remark    string               `json:"remark,omitempty"`
	CreatedAt string               `json:"created_at,omitempty"`
	UpdatedAt string               `json:"updated_at,omitempty"`
	Children  []DepartmentTreeItem `json:"children,omitempty"`
}

// MenuTreeItem 前端使用的菜单树形结构
type MenuTreeItem struct {
	ID         uint           `json:"id"`
	ParentID   uint           `json:"parent_id"`
	Name       string         `json:"name"`
	Title      string         `json:"title,omitempty"`
	Slug       string         `json:"slug"`
	Icon       string         `json:"icon,omitempty"`
	Path       string         `json:"path,omitempty"`
	Component  string         `json:"component,omitempty"`
	Permission string         `json:"permission,omitempty"`
	Type       uint8          `json:"type"`
	Status     uint8          `json:"status"`
	Sort       int            `json:"sort"`
	IsHidden   uint8          `json:"is_hidden"`
	LinkType   uint8          `json:"link_type"`
	OpenType   uint8          `json:"open_type"`
	NoCache    uint8          `json:"no_cache"`
	CreatedAt  string         `json:"created_at,omitempty"`
	UpdatedAt  string         `json:"updated_at,omitempty"`
	Children   []MenuTreeItem `json:"children,omitempty"`
}

// ConvertDepartmentTree 将 Department 树形结构转换为前端格式
func ConvertDepartmentTree(departments []models.Department) []DepartmentTreeItem {
	result := make([]DepartmentTreeItem, 0, len(departments))
	for _, dept := range departments {
		item := DepartmentTreeItem{
			ID:       dept.ID,
			ParentID: dept.ParentID,
			Name:     dept.Name,
			Code:     dept.Code,
			Leader:   dept.Leader,
			Phone:    dept.Phone,
			Email:    dept.Email,
			Status:   dept.Status,
			Sort:     dept.Sort,
			Remark:   dept.Remark,
		}

		if dept.CreatedAt != nil && !dept.CreatedAt.IsZero() {
			item.CreatedAt = dept.CreatedAt.ToDateTimeString()
		}
		if dept.UpdatedAt != nil && !dept.UpdatedAt.IsZero() {
			item.UpdatedAt = dept.UpdatedAt.ToDateTimeString()
		}

		// 递归转换子节点
		if len(dept.Children) > 0 {
			item.Children = ConvertDepartmentTree(dept.Children)
		}

		result = append(result, item)
	}
	return result
}

// ConvertMenuTree 将 Menu 树形结构转换为前端格式
func ConvertMenuTree(menus []models.Menu) []MenuTreeItem {
	result := make([]MenuTreeItem, 0, len(menus))
	for _, menu := range menus {
		item := MenuTreeItem{
			ID:         menu.ID,
			ParentID:   menu.ParentID,
			Name:       menu.Title, // 前端使用 name 字段，对应后端的 Title
			Title:      menu.Title,
			Slug:       menu.Slug,
			Icon:       menu.Icon,
			Path:       menu.Path,
			Component:  menu.Component,
			Permission: menu.Permission,
			Type:       menu.Type,
			Status:     menu.Status,
			Sort:       menu.Sort,
			IsHidden:   menu.IsHidden,
			LinkType:   menu.LinkType,
			OpenType:   menu.OpenType,
			NoCache:    menu.NoCache,
		}

		if menu.CreatedAt != nil && !menu.CreatedAt.IsZero() {
			item.CreatedAt = menu.CreatedAt.ToDateTimeString()
		}
		if menu.UpdatedAt != nil && !menu.UpdatedAt.IsZero() {
			item.UpdatedAt = menu.UpdatedAt.ToDateTimeString()
		}

		// 递归转换子节点
		if len(menu.Children) > 0 {
			item.Children = ConvertMenuTree(menu.Children)
		}

		result = append(result, item)
	}
	return result
}
