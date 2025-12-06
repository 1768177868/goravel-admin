package jobs

import (
	"github.com/goravel/framework/facades"
)

// ProcessImage 处理图片任务
type ProcessImage struct {
}

func (r *ProcessImage) Signature() string {
	return "process_image"
}

func (r *ProcessImage) Handle(args ...any) error {
	if len(args) > 0 {
		imagePath := args[0].(string)

		facades.Log().Infof("🖼️ [Job] 处理图片 - 路径: %s", imagePath)
		// 实际场景中这里会进行图片压缩、裁剪、生成缩略图等操作
	}
	return nil
}
