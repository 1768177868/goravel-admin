package services

import (
	"context"
	"fmt"

	"github.com/goravel/framework/facades"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type ExportRecordService interface {
	// GetByID 根据ID获取导出记录
	GetByID(id uint) (*models.Export, error)
	// GetByIDs 根据ID列表获取导出记录
	GetByIDs(ids []uint) ([]models.Export, error)
	// GetList 获取导出记录列表
	GetList(filters ExportRecordFilters, page, pageSize int) ([]models.Export, int64, error)
	// Delete 删除导出记录
	Delete(id uint) error
	// BatchDelete 批量删除导出记录
	BatchDelete(ids []uint) error
}

// ExportRecordFilters 导出记录查询过滤器
type ExportRecordFilters struct {
	AdminID   string
	Filename  string
	Disk      string
	Status    string
	StartTime string
	EndTime   string
	OrderBy   string
}

type ExportRecordServiceImpl struct {
}

func NewExportRecordService() ExportRecordService {
	return &ExportRecordServiceImpl{}
}

// GetByID 根据ID获取导出记录
func (s *ExportRecordServiceImpl) GetByID(id uint) (*models.Export, error) {
	var export models.Export
	if err := facades.Orm().Query().Where("id", id).First(&export); err != nil {
		return nil, apperrors.ErrExportRecordNotFound.WithError(err)
	}
	return &export, nil
}

// GetByIDs 根据ID列表获取导出记录
func (s *ExportRecordServiceImpl) GetByIDs(ids []uint) ([]models.Export, error) {
	if len(ids) == 0 {
		return []models.Export{}, nil
	}

	idsAny := helpers.ConvertUintSliceToAny(ids)
	var exports []models.Export
	if err := facades.Orm().Query().WhereIn("id", idsAny).Get(&exports); err != nil {
		return nil, apperrors.ErrQueryFailed.WithError(err)
	}
	return exports, nil
}

// GetList 获取导出记录列表
func (s *ExportRecordServiceImpl) GetList(filters ExportRecordFilters, page, pageSize int) ([]models.Export, int64, error) {
	query := facades.Orm().Query().Model(&models.Export{})

	// 应用筛选条件
	if filters.AdminID != "" {
		query = query.Where("admin_id", filters.AdminID)
	}
	if filters.Filename != "" {
		query = query.Where("filename LIKE ?", "%"+filters.Filename+"%")
	}
	if filters.Disk != "" {
		query = query.Where("disk = ?", filters.Disk)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.StartTime != "" {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if filters.EndTime != "" {
		query = query.Where("created_at <= ?", filters.EndTime)
	}

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "id:desc"
	}
	query = helpers.ApplySort(query, orderBy, "id:desc")

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	var exports []models.Export
	err = query.With("Admin").
		Offset((page-1)*pageSize).
		Limit(pageSize).
		Get(&exports)
	if err != nil {
		return nil, 0, err
	}

	return exports, total, nil
}

// Delete 删除导出记录
func (s *ExportRecordServiceImpl) Delete(id uint) error {
	var export models.Export
	if err := facades.Orm().Query().Where("id", id).First(&export); err != nil {
		return fmt.Errorf("导出记录不存在: %v", err)
	}

	if _, err := facades.Orm().Query().Delete(&export); err != nil {
		errorlog.Record(context.Background(), "export-record", "删除导出记录失败", map[string]any{
			"export_id": id,
			"error":     err.Error(),
		}, "删除导出记录失败: %v", err)
		return apperrors.ErrDeleteFailed.WithError(err)
	}

	return nil
}

// BatchDelete 批量删除导出记录
func (s *ExportRecordServiceImpl) BatchDelete(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	idsAny := helpers.ConvertUintSliceToAny(ids)
	if _, err := facades.Orm().Query().WhereIn("id", idsAny).Delete(&models.Export{}); err != nil {
		errorlog.Record(context.Background(), "export-record", "批量删除导出记录失败", map[string]any{
			"ids":   ids,
			"count": len(ids),
			"error": err.Error(),
		}, "批量删除导出记录失败: %v", err)
		return apperrors.ErrBatchDeleteExportFailed.WithError(err)
	}

	return nil
}

