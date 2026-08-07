package response

import (
	stdhttp "net/http"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/contracts/http"

	apperrors "goravel/app/errors"
	"goravel/app/utils"
)

// OpenStorageDisk 安全打开磁盘驱动；失败时返回可直接响应的 http.Response。
func OpenStorageDisk(ctx http.Context, module, disk string, attrs map[string]any) (filesystem.Driver, http.Response, bool) {
	storage, err := utils.StorageDisk(disk)
	if err != nil {
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return nil, Error(ctx, stdhttp.StatusBadRequest, businessErr), false
		}
		if attrs == nil {
			attrs = map[string]any{"disk": disk}
		} else if _, ok := attrs["disk"]; !ok {
			attrs["disk"] = disk
		}
		return nil, ErrorWithLog(ctx, module, err, attrs), false
	}
	return storage, nil, true
}
