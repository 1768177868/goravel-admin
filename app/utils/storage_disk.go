package utils

import (
	"fmt"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	fwfilesystem "github.com/goravel/framework/filesystem"

	apperrors "goravel/app/errors"
)

// ValidateFilesystemDisk 检查云存储磁盘的必填配置是否已写入 .env / config。
// 未配置时返回业务错误，避免 Storage().Disk() 内部 panic。
func ValidateFilesystemDisk(disk string) error {
	switch disk {
	case "", "local", "public":
		return nil
	case "s3":
		return requireDiskFields(disk, "key", "secret", "region", "bucket", "url")
	case "oss":
		return requireDiskFields(disk, "key", "secret", "bucket", "url", "endpoint")
	case "cos":
		return requireDiskFields(disk, "key", "secret", "url")
	case "minio":
		return requireDiskFields(disk, "key", "secret", "bucket", "url", "endpoint")
	default:
		return storageDiskNotConfigured(disk)
	}
}

func requireDiskFields(disk string, fields ...string) error {
	cfg := facades.Config()
	for _, field := range fields {
		if cfg.GetString(fmt.Sprintf("filesystems.disks.%s.%s", disk, field)) == "" {
			return storageDiskNotConfigured(disk)
		}
	}
	return nil
}

func storageDiskNotConfigured(disk string) *apperrors.BusinessError {
	return apperrors.NewBusinessError(
		apperrors.ErrStorageDiskNotConfigured.Code,
		apperrors.ErrStorageDiskNotConfigured.Message,
	).WithParams(map[string]any{"disk": disk})
}

// StorageDisk 安全获取磁盘驱动：先校验配置，再用 NewDriver，避免 Disk() panic。
func StorageDisk(disk string) (filesystem.Driver, error) {
	if disk == "" {
		disk = facades.Config().GetString("filesystems.default", "local")
	}
	if err := ValidateFilesystemDisk(disk); err != nil {
		return nil, err
	}
	driver, err := fwfilesystem.NewDriver(facades.Config(), disk)
	if err != nil {
		return nil, storageDiskNotConfigured(disk).WithError(err)
	}
	return driver, nil
}
