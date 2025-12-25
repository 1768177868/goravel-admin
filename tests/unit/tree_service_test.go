package unit

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TreeServiceTestSuite 树形服务测试套件
type TreeServiceTestSuite struct {
	suite.Suite
}

func TestTreeServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TreeServiceTestSuite))
}

func (s *TreeServiceTestSuite) SetupTest() {
}

func (s *TreeServiceTestSuite) TearDownTest() {
}

// TreeNode 测试用树形节点
type TreeNode struct {
	ID       uint
	ParentID uint
	Name     string
	Children []TreeNode
}

// buildTree 构建树形结构
func buildTree(nodes []TreeNode) []TreeNode {
	if len(nodes) == 0 {
		return []TreeNode{}
	}

	// 先创建所有节点的映射
	nodeMap := make(map[uint]*TreeNode)
	for i := range nodes {
		nodes[i].Children = []TreeNode{}
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	// 构建父子关系
	for i := range nodes {
		if nodes[i].ParentID != 0 {
			if parent, ok := nodeMap[nodes[i].ParentID]; ok {
				parent.Children = append(parent.Children, nodes[i])
			}
		}
	}

	// 收集根节点（从 map 中获取，确保 Children 已更新）
	var roots []TreeNode
	for i := range nodes {
		if nodes[i].ParentID == 0 {
			roots = append(roots, *nodeMap[nodes[i].ID])
		}
	}

	return roots
}

// flattenTree 展平树形结构
func flattenTree(nodes []TreeNode) []TreeNode {
	var result []TreeNode
	for _, node := range nodes {
		result = append(result, node)
		if len(node.Children) > 0 {
			result = append(result, flattenTree(node.Children)...)
		}
	}
	return result
}

// contains 检查切片是否包含元素
func contains(slice []uint, item uint) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// uniqueUint 去重
func uniqueUint(slice []uint) []uint {
	seen := make(map[uint]bool)
	var result []uint
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// TestBuildTree_EmptyList 测试空列表
func (s *TreeServiceTestSuite) TestBuildTree_EmptyList() {
	result := buildTree([]TreeNode{})
	s.Empty(result)
}

// TestBuildTree_SingleRoot 测试单个根节点
func (s *TreeServiceTestSuite) TestBuildTree_SingleRoot() {
	nodes := []TreeNode{
		{ID: 1, ParentID: 0, Name: "Root"},
	}
	result := buildTree(nodes)

	s.Len(result, 1)
	s.Equal("Root", result[0].Name)
}

// TestBuildTree_MultipleRoots 测试多个根节点
func (s *TreeServiceTestSuite) TestBuildTree_MultipleRoots() {
	nodes := []TreeNode{
		{ID: 1, ParentID: 0, Name: "Root1"},
		{ID: 2, ParentID: 0, Name: "Root2"},
		{ID: 3, ParentID: 0, Name: "Root3"},
	}
	result := buildTree(nodes)

	s.Len(result, 3)
}

// TestBuildTree_ParentChild 测试父子关系
func (s *TreeServiceTestSuite) TestBuildTree_ParentChild() {
	nodes := []TreeNode{
		{ID: 1, ParentID: 0, Name: "Root"},
		{ID: 2, ParentID: 1, Name: "Child1"},
		{ID: 3, ParentID: 1, Name: "Child2"},
		{ID: 4, ParentID: 2, Name: "Grandchild"},
	}
	result := buildTree(nodes)

	s.Len(result, 1, "应该只有 1 个根节点")
	s.Len(result[0].Children, 2, "根节点应该有 2 个子节点")
}

// TestFlattenTree 测试树形展平
func (s *TreeServiceTestSuite) TestFlattenTree() {
	tree := []TreeNode{
		{
			ID:   1,
			Name: "Root",
			Children: []TreeNode{
				{ID: 2, Name: "Child1"},
				{ID: 3, Name: "Child2"},
			},
		},
	}
	result := flattenTree(tree)

	s.Len(result, 3)
	s.Equal("Root", result[0].Name)
	s.Equal("Child1", result[1].Name)
	s.Equal("Child2", result[2].Name)
}

// TestContains 测试 contains 函数
func (s *TreeServiceTestSuite) TestContains() {
	tests := []struct {
		name  string
		slice []uint
		item  uint
		want  bool
	}{
		{"空切片", []uint{}, 1, false},
		{"存在的元素", []uint{1, 2, 3}, 2, true},
		{"不存在的元素", []uint{1, 2, 3}, 4, false},
		{"单元素存在", []uint{5}, 5, true},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, contains(tt.slice, tt.item))
		})
	}
}

// TestUniqueUint 测试去重函数
func (s *TreeServiceTestSuite) TestUniqueUint() {
	tests := []struct {
		name  string
		slice []uint
		want  int // 期望的长度
	}{
		{"空切片", []uint{}, 0},
		{"无重复", []uint{1, 2, 3}, 3},
		{"有重复", []uint{1, 2, 2, 3, 3, 3}, 3},
		{"全部重复", []uint{5, 5, 5}, 1},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := uniqueUint(tt.slice)
			s.Len(got, tt.want)
		})
	}
}
