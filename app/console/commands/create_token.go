package commands

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/services"
)

// # Admin用户 - 1小时后过期
// go run . artisan token:create 1 --expires=1h

// # 7天后过期
// go run . artisan token:create 1 --expires=7d

// # 使用简写
// go run . artisan token:create 1 -e=1h

// # 指定token名称
// go run . artisan token:create 1 --expires=7d --name=api-token

// # 创建永久token（不设置expires）
// go run . artisan token:create 1

type CreateToken struct {
}

// Signature The name and signature of the console command.
func (receiver *CreateToken) Signature() string {
	return "token:create"
}

// Description The console command description.
func (receiver *CreateToken) Description() string {
	return "为指定管理员创建token（永久或指定过期时间）"
}

// Extend The console command extend.
func (receiver *CreateToken) Extend() command.Extend {
	return command.Extend{
		Category: "token",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "expires",
				Aliases: []string{"e"},
				Usage:   "过期时间，格式：1h（1小时）、24h（24小时）、7d（7天）等，不设置则创建永久token",
			},
			&command.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Value:   "console-token",
				Usage:   "token名称",
			},
		},
	}
}

// Handle Execute the console command.
func (receiver *CreateToken) Handle(ctx console.Context) error {
	// 获取用户标识（ID或用户名）
	userIdentifier := ctx.Argument(0)
	if userIdentifier == "" {
		return errors.New("请提供用户ID或用户名")
	}

	// 用户类型固定为 admin
	userType := "admin"

	// 获取过期时间选项
	expiresStr := ctx.Option("expires")
	var expiresAt *time.Time

	if expiresStr != "" {
		// 解析过期时间
		duration, err := parseDuration(expiresStr)
		if err != nil {
			return errors.New("过期时间格式错误，请使用：1h（1小时）、24h（24小时）、7d（7天）等格式")
		}
		exp := time.Now().Add(duration)
		expiresAt = &exp
	}
	// 如果 expiresStr 为空，expiresAt 为 nil，表示永久token

	// 获取token名称
	tokenName := ctx.Option("name")
	if tokenName == "" {
		tokenName = "console-token"
	}

	// 查询管理员
	var admin models.Admin
	var err error
	// 尝试作为ID解析
	if id, parseErr := strconv.ParseUint(userIdentifier, 10, 32); parseErr == nil {
		// 作为ID查询
		err = facades.Orm().Query().Where("id", uint(id)).First(&admin)
	} else {
		// 作为用户名查询
		err = facades.Orm().Query().Where("username", userIdentifier).First(&admin)
	}

	if err != nil {
		return errors.New("管理员不存在")
	}

	userID := admin.ID
	username := admin.Username

	// 创建token
	tokenService := services.NewTokenServiceImpl()
	// 命令行创建token时，浏览器、IP、操作系统信息为空
	plainToken, accessToken, err := tokenService.CreateToken(userType, userID, tokenName, expiresAt, "", "", "", "")
	if err != nil {
		return errors.New("创建token失败: " + err.Error())
	}

	// 输出结果
	ctx.Info("Token创建成功！")
	ctx.Line("")
	ctx.Info("用户信息:")
	ctx.Line("  类型: " + userType)
	ctx.Line("  ID: " + strconv.FormatUint(uint64(userID), 10))
	ctx.Line("  用户名: " + username)
	ctx.Line("")
	ctx.Info("Token信息:")
	ctx.Line("  名称: " + accessToken.Name)
	if accessToken.ExpiresAt != nil {
		ctx.Line("  过期时间: " + accessToken.ExpiresAt.Format("2006-01-02 15:04:05"))
	} else {
		ctx.Line("  过期时间: 永久有效")
	}
	ctx.Line("  创建时间: " + accessToken.CreatedAt.Format("2006-01-02 15:04:05"))
	ctx.Line("")
	ctx.Warning("请妥善保管以下token，它只会显示一次：")
	ctx.Line("")
	ctx.Line(plainToken)
	ctx.Line("")

	return nil
}

// parseDuration 解析时间字符串，支持 h（小时）、d（天）、m（分钟）等格式
func parseDuration(s string) (time.Duration, error) {
	// 移除空格
	s = strings.TrimSpace(s)

	if len(s) == 0 {
		return 0, errors.New("时间字符串为空")
	}

	// 获取最后一个字符作为单位
	unit := s[len(s)-1:]
	valueStr := s[:len(s)-1]

	// 解析数值
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, err
	}

	// 根据单位转换为duration
	switch unit {
	case "m", "M":
		// 分钟
		return time.Duration(value) * time.Minute, nil
	case "h", "H":
		// 小时
		return time.Duration(value) * time.Hour, nil
	case "d", "D":
		// 天
		return time.Duration(value) * 24 * time.Hour, nil
	case "w", "W":
		// 周
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	default:
		// 尝试直接解析为duration（如 "1h30m"）
		return time.ParseDuration(s)
	}
}
