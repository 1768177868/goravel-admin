package jobs

import (
	"github.com/goravel/framework/facades"

	"goravel/app/errors"
)

// ProcessImageArgs 处理图片任务的参数结构体
type ProcessImageArgs struct {
	ImagePath string `validate:"required"`
}

// ProcessImage 处理图片任务
type ProcessImage struct {
}

func (r *ProcessImage) Signature() string {
	return "process_image"
}

// Handle 处理图片任务
// 
// 参数:
//   - args[0]: ProcessImageArgs 结构体或 string (图片路径)
//
// 返回:
//   - error: 错误信息
func (r *ProcessImage) Handle(args ...any) error {
	if len(args) < 1 {
		return errors.ErrInvalidArgument.WithMessage("missing image path")
	}

	var imagePath string
	switch v := args[0].(type) {
	case ProcessImageArgs:
		imagePath = v.ImagePath
	case string:
		imagePath = v
	case map[string]any:
		if path, ok := v["image_path"].(string); ok {
			imagePath = path
		} else if path, ok := v["path"].(string); ok {
			imagePath = path
		}
	default:
		return errors.ErrInvalidArgument.WithMessage("invalid argument type for image path")
	}

	if imagePath == "" {
		return errors.ErrInvalidArgument.WithMessage("image path is required")
	}

	facades.Log().Infof("🖼️ [Job] 处理图片 - 路径: %s", imagePath)
	// 实际场景中这里会进行图片压缩、裁剪、生成缩略图等操作
	return nil
}
