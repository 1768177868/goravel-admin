package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/path"

	"goravel/app/utils"
)

type ClearChunks struct {
}

// Signature The name and signature of the console command.
func (r *ClearChunks) Signature() string {
	return "app:clear-chunks"
}

// Description The console command description.
func (r *ClearChunks) Description() string {
	return "清理3天前的分片文件"
}

// Extend The console command extend.
func (r *ClearChunks) Extend() command.Extend {
	return command.Extend{Category: "app"}
}

// Handle Execute the console command.
func (r *ClearChunks) Handle(ctx console.Context) error {
	// 从数据库读取文件存储配置
	disk := utils.GetConfigValue("storage", "file_disk", "")
	if disk == "" {
		disk = "local"
	}

	// 检查存储驱动是否为本地存储
	if disk != "local" && disk != "public" {
		ctx.Info(fmt.Sprintf("当前存储驱动为 %s，清理分片文件功能仅支持本地存储，跳过清理", disk))
		return nil
	}

	storage := facades.Storage().Disk(disk)

	// 计算3天前的时间
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	ctx.Info(fmt.Sprintf("开始清理3天前的分片文件（%s之前创建的文件）...", threeDaysAgo.Format("2006-01-02 15:04:05")))

	// 获取存储根目录
	var storageRoot string
	if disk == "public" {
		storageRoot = path.Storage("app/public")
	} else {
		storageRoot = path.Storage("app")
	}

	chunksDir := filepath.Join(storageRoot, "chunks")

	// 检查目录是否存在
	if _, err := os.Stat(chunksDir); os.IsNotExist(err) {
		ctx.Info("分片目录不存在，无需清理")
		return nil
	}

	cleanedCount := 0
	cleanedSize := int64(0)
	errorCount := 0

	// 遍历 chunks 目录下的所有文件
	err := filepath.Walk(chunksDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 如果无法访问某个文件/目录，记录错误但继续
			ctx.Info(fmt.Sprintf("无法访问路径 %s: %v", path, err))
			errorCount++
			return nil
		}

		// 跳过根目录本身
		if path == chunksDir {
			return nil
		}

		// 只处理文件（分片文件）
		if !info.IsDir() {
			// 跳过 .gitignore 文件，避免误删
			if info.Name() == ".gitignore" {
				return nil
			}

			// 检查文件修改时间
			if info.ModTime().Before(threeDaysAgo) {
				// 使用 Storage 接口删除文件（保持一致性）
				relativePath := strings.TrimPrefix(path, storageRoot+string(filepath.Separator))
				relativePath = strings.ReplaceAll(relativePath, string(filepath.Separator), "/")

				if err := storage.Delete(relativePath); err != nil {
					ctx.Info(fmt.Sprintf("删除分片文件失败 %s: %v", relativePath, err))
					errorCount++
				} else {
					cleanedCount++
					cleanedSize += info.Size()
				}
			}
		}

		return nil
	})

	if err != nil {
		ctx.Error(fmt.Sprintf("遍历分片目录失败: %v", err))
		return err
	}

	// 清理空目录（避免留下大量空目录）
	emptyDirCount := r.cleanupEmptyDirs(chunksDir, ctx)

	ctx.Info(fmt.Sprintf("分片文件清理完成！已清理 %d 个文件，释放空间 %s，删除 %d 个空目录，错误数: %d",
		cleanedCount,
		r.formatFileSize(cleanedSize),
		emptyDirCount,
		errorCount))

	return nil
}

// cleanupEmptyDirs 清理空目录
func (r *ClearChunks) cleanupEmptyDirs(rootDir string, ctx console.Context) int {
	var dirs []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && path != rootDir {
			dirs = append(dirs, path)
		}
		return nil
	})

	if err != nil {
		return 0
	}

	// 反向遍历目录（从最深层的开始）
	removedCount := 0
	for i := len(dirs) - 1; i >= 0; i-- {
		dir := dirs[i]
		// 检查目录是否为空
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(entries) == 0 {
			// 目录为空，删除它
			if err := os.Remove(dir); err == nil {
				removedCount++
			}
		}
	}

	return removedCount
}

// formatFileSize 格式化文件大小
func (r *ClearChunks) formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
