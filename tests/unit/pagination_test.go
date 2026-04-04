package unit

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/http/helpers"
)

// PaginationTestSuite 分页测试套件
type PaginationTestSuite struct {
	suite.Suite
}

func TestPaginationTestSuite(t *testing.T) {
	suite.Run(t, new(PaginationTestSuite))
}

func (s *PaginationTestSuite) SetupTest() {
}

func (s *PaginationTestSuite) TearDownTest() {
}

// paginateSlice 模拟分页函数
func paginateSlice[T any](slice []T, page, pageSize int) ([]T, int64) {
	total := int64(len(slice))
	if total == 0 {
		return []T{}, 0
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(slice) {
		return []T{}, total
	}

	if end > len(slice) {
		end = len(slice)
	}

	return slice[start:end], total
}

// TestPaginateSlice_EmptySlice 测试空切片
func (s *PaginationTestSuite) TestPaginateSlice_EmptySlice() {
	result, total := paginateSlice([]int{}, 1, 10)

	s.Empty(result)
	s.Equal(int64(0), total)
}

// TestPaginateSlice_FirstPage 测试第一页
func (s *PaginationTestSuite) TestPaginateSlice_FirstPage() {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result, total := paginateSlice(slice, 1, 3)

	s.Equal([]int{1, 2, 3}, result)
	s.Equal(int64(10), total)
}

// TestPaginateSlice_MiddlePage 测试中间页
func (s *PaginationTestSuite) TestPaginateSlice_MiddlePage() {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result, total := paginateSlice(slice, 2, 3)

	s.Equal([]int{4, 5, 6}, result)
	s.Equal(int64(10), total)
}

// TestPaginateSlice_LastPage 测试最后一页（不满）
func (s *PaginationTestSuite) TestPaginateSlice_LastPage() {
	slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result, total := paginateSlice(slice, 4, 3)

	s.Equal([]int{10}, result)
	s.Equal(int64(10), total)
}

// TestPaginateSlice_OutOfRange 测试超出范围的页码
func (s *PaginationTestSuite) TestPaginateSlice_OutOfRange() {
	slice := []int{1, 2, 3, 4, 5}
	result, total := paginateSlice(slice, 10, 3)

	s.Empty(result)
	s.Equal(int64(5), total)
}

// TestValidatePagination 测试分页参数验证（默认边界）
func (s *PaginationTestSuite) TestValidatePagination() {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		wantPage     int
		wantPageSize int
	}{
		{"正常值", 1, 10, 1, 10},
		{"page 小于 1", 0, 10, 1, 10},
		{"page 为负数", -5, 10, 1, 10},
		{"pageSize 小于 1", 1, 0, 1, 10},
		{"pageSize 超过最大值", 1, 200, 1, 100},
		{"两者都异常", -1, -1, 1, 10},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			gotPage, gotPageSize := helpers.ValidatePagination(tt.page, tt.pageSize)
			s.Equal(tt.wantPage, gotPage)
			s.Equal(tt.wantPageSize, gotPageSize)
		})
	}
}

func (s *PaginationTestSuite) TestValidatePaginationEx_CustomMax() {
	p, ps := helpers.ValidatePaginationEx(1, 80, helpers.PaginationLimits{DefaultPageSize: 20, MaxPageSize: 50})
	s.Equal(1, p)
	s.Equal(50, ps)

	p, ps = helpers.ValidatePaginationEx(1, 0, helpers.PaginationLimits{DefaultPageSize: 20, MaxPageSize: 50})
	s.Equal(1, p)
	s.Equal(20, ps)
}

func (s *PaginationTestSuite) TestTotalPages() {
	s.Equal(0, helpers.TotalPages(0, 10))
	s.Equal(0, helpers.TotalPages(5, 0))
	s.Equal(1, helpers.TotalPages(1, 10))
	s.Equal(2, helpers.TotalPages(11, 10))
	s.Equal(3, helpers.TotalPages(21, 10))
}
