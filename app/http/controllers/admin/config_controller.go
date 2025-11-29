package admin

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/spf13/cast"

	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type ConfigController struct {
}

func NewConfigController() *ConfigController {
	return &ConfigController{}
}

// GetByGroup 根据分组获取配置
func (r *ConfigController) GetByGroup(ctx http.Context) http.Response {
	group := ctx.Request().Route("group")
	if group == "" {
		return response.Error(ctx, http.StatusBadRequest, "config_group_required")
	}

	var configs []models.Config
	// 查询配置，即使没有数据也返回空数组，不返回错误
	_ = facades.Orm().Query().Where("group", group).Order("sort asc, id asc").Get(&configs)

	return response.Success(ctx, "get_success", http.Json{
		"configs": configs,
	})
}

// Save 保存配置（按分组批量保存）
func (r *ConfigController) Save(ctx http.Context) http.Response {
	group := ctx.Request().Input("group")
	if group == "" {
		return response.Error(ctx, http.StatusBadRequest, "config_group_required")
	}

	configsMap := ctx.Request().InputMap("configs")
	if len(configsMap) == 0 {
		return response.Error(ctx, http.StatusBadRequest, "configs_required")
	}

	// 获取该分组下的所有配置（即使查询失败也继续，使用空数组）
	var existingConfigs []models.Config
	_ = facades.Orm().Query().Where("group", group).Get(&existingConfigs)

	// 创建key到config的映射
	configMap := make(map[string]*models.Config)
	for i := range existingConfigs {
		configMap[existingConfigs[i].Key] = &existingConfigs[i]
	}

	now := carbon.Now()

	// 批量处理配置更新和创建
	for key, value := range configsMap {
		// 转换值为字符串，处理布尔值
		var valueStr string
		switch v := value.(type) {
		case bool:
			if v {
				valueStr = "1"
			} else {
				valueStr = "0"
			}
		case nil:
			valueStr = ""
		default:
			valueStr = cast.ToString(value)
		}

		if config, exists := configMap[key]; exists {
			// 更新现有配置
			config.Value = valueStr
			if err := facades.Orm().Query().Save(config); err != nil {
				errorlog.RecordHTTP(ctx, "config", "Failed to update config", map[string]any{
					"error": err.Error(),
					"group": group,
					"key":   key,
				}, "Update config error: %v", err)
				return response.Error(ctx, http.StatusInternalServerError, "update_failed")
			}
		} else {
			// 创建新配置
			configData := map[string]interface{}{
				"group":      group,
				"key":        key,
				"value":      valueStr,
				"type":       "input",
				"sort":       0,
				"created_at": now,
				"updated_at": now,
			}
			if err := facades.Orm().Query().Table("configs").Create(configData); err != nil {
				errorlog.RecordHTTP(ctx, "config", "Failed to create config", map[string]any{
					"error": err.Error(),
					"group": group,
					"key":   key,
				}, "Create config error: %v", err)
				return response.Error(ctx, http.StatusInternalServerError, "create_failed")
			}
		}
	}

	return response.Success(ctx, "save_success")
}

// TestEmail 测试邮件发送
func (r *ConfigController) TestEmail(ctx http.Context) http.Response {
	emailHost := ctx.Request().Input("email_host")
	emailPort := cast.ToInt(ctx.Request().Input("email_port", "587"))
	emailUsername := ctx.Request().Input("email_username")
	emailPassword := ctx.Request().Input("email_password")
	emailFrom := ctx.Request().Input("email_from")
	emailFromName := ctx.Request().Input("email_from_name")
	emailEncryption := ctx.Request().Input("email_encryption", "tls")

	// 验证必填字段
	if emailHost == "" || emailPort == 0 || emailUsername == "" || emailFrom == "" {
		return response.Error(ctx, http.StatusBadRequest, "email_config_required")
	}

	// 获取当前登录的管理员邮箱作为测试收件人
	adminValue := ctx.Value("admin")
	if adminValue == nil {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	admin, ok := adminValue.(models.Admin)
	if !ok {
		return response.Error(ctx, http.StatusUnauthorized, "not_logged_in")
	}

	// 如果没有邮箱，使用发件人邮箱作为测试收件人
	testEmail := emailFrom
	if admin.Email != "" {
		testEmail = admin.Email
	}

	// 构建邮件内容
	fromName := emailFromName
	if fromName == "" {
		fromName = emailFrom
	}
	subject := "测试邮件"
	body := fmt.Sprintf(`<h2>这是一封测试邮件</h2>
<p>如果您收到这封邮件，说明邮件配置正确。</p>
<p>发送时间：%s</p>
<p>SMTP服务器：%s:%d</p>
<p>加密方式：%s</p>`, carbon.Now().ToDateTimeString(), emailHost, emailPort, emailEncryption)

	// 构建邮件消息
	message := fmt.Sprintf("From: %s <%s>\r\n", fromName, emailFrom)
	message += fmt.Sprintf("To: %s\r\n", testEmail)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/html; charset=UTF-8\r\n"
	message += "\r\n" + body

	// 构建SMTP地址
	addr := fmt.Sprintf("%s:%d", emailHost, emailPort)

	// 创建SMTP认证
	auth := smtp.PlainAuth("", emailUsername, emailPassword, emailHost)

	// 发送邮件
	var err error
	if emailEncryption == "ssl" {
		// SSL连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         emailHost,
		}
		conn, connErr := tls.Dial("tcp", addr, tlsConfig)
		if connErr != nil {
			errorlog.RecordHTTP(ctx, "config", "Failed to connect SMTP server", map[string]any{
				"error": connErr.Error(),
				"host":  emailHost,
				"port":  emailPort,
			}, "SMTP SSL connection error: %v", connErr)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_connection_failed")
		}
		defer conn.Close()

		client, clientErr := smtp.NewClient(conn, emailHost)
		if clientErr != nil {
			errorlog.RecordHTTP(ctx, "config", "Failed to create SMTP client", map[string]any{
				"error": clientErr.Error(),
			}, "SMTP client creation error: %v", clientErr)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_client_failed")
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			errorlog.RecordHTTP(ctx, "config", "Failed to authenticate SMTP", map[string]any{
				"error": err.Error(),
			}, "SMTP authentication error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_auth_failed")
		}

		if err = client.Mail(emailFrom); err != nil {
			errorlog.RecordHTTP(ctx, "config", "Failed to set sender", map[string]any{
				"error": err.Error(),
			}, "SMTP set sender error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_sender_failed")
		}

		if err = client.Rcpt(testEmail); err != nil {
			errorlog.RecordHTTP(ctx, "config", "Failed to set recipient", map[string]any{
				"error": err.Error(),
			}, "SMTP set recipient error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_recipient_failed")
		}

		writer, writerErr := client.Data()
		if writerErr != nil {
			errorlog.RecordHTTP(ctx, "config", "Failed to get data writer", map[string]any{
				"error": writerErr.Error(),
			}, "SMTP data writer error: %v", writerErr)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_data_failed")
		}

		_, err = writer.Write([]byte(message))
		if err != nil {
			writer.Close()
			errorlog.RecordHTTP(ctx, "config", "Failed to write email data", map[string]any{
				"error": err.Error(),
			}, "SMTP write data error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_write_failed")
		}

		err = writer.Close()
		if err != nil {
			errorlog.RecordHTTP(ctx, "config", "Failed to close data writer", map[string]any{
				"error": err.Error(),
			}, "SMTP close writer error: %v", err)
			return response.Error(ctx, http.StatusInternalServerError, "test_email_close_failed")
		}
	} else {
		// TLS或普通连接
		if emailEncryption == "tls" {
			// TLS连接
			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
				ServerName:         emailHost,
			}
			err = smtp.SendMail(addr, auth, emailFrom, []string{testEmail}, []byte(message))
			if err != nil {
				// 如果直接SendMail失败，尝试手动TLS
				conn, connErr := smtp.Dial(addr)
				if connErr != nil {
					errorlog.RecordHTTP(ctx, "config", "Failed to dial SMTP server", map[string]any{
						"error": connErr.Error(),
						"host":  emailHost,
						"port":  emailPort,
					}, "SMTP dial error: %v", connErr)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_connection_failed")
				}
				defer conn.Close()

				if err = conn.StartTLS(tlsConfig); err != nil {
					errorlog.RecordHTTP(ctx, "config", "Failed to start TLS", map[string]any{
						"error": err.Error(),
					}, "SMTP StartTLS error: %v", err)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_tls_failed")
				}

				if err = conn.Auth(auth); err != nil {
					errorlog.RecordHTTP(ctx, "config", "Failed to authenticate SMTP", map[string]any{
						"error": err.Error(),
					}, "SMTP authentication error: %v", err)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_auth_failed")
				}

				if err = conn.Mail(emailFrom); err != nil {
					errorlog.RecordHTTP(ctx, "config", "Failed to set sender", map[string]any{
						"error": err.Error(),
					}, "SMTP set sender error: %v", err)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_sender_failed")
				}

				if err = conn.Rcpt(testEmail); err != nil {
					errorlog.RecordHTTP(ctx, "config", "Failed to set recipient", map[string]any{
						"error": err.Error(),
					}, "SMTP set recipient error: %v", err)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_recipient_failed")
				}

				writer, writerErr := conn.Data()
				if writerErr != nil {
					errorlog.RecordHTTP(ctx, "config", "Failed to get data writer", map[string]any{
						"error": writerErr.Error(),
					}, "SMTP data writer error: %v", writerErr)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_data_failed")
				}

				_, err = writer.Write([]byte(message))
				if err != nil {
					writer.Close()
					errorlog.RecordHTTP(ctx, "config", "Failed to write email data", map[string]any{
						"error": err.Error(),
					}, "SMTP write data error: %v", err)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_write_failed")
				}

				err = writer.Close()
				if err != nil {
					errorlog.RecordHTTP(ctx, "config", "Failed to close data writer", map[string]any{
						"error": err.Error(),
					}, "SMTP close writer error: %v", err)
					return response.Error(ctx, http.StatusInternalServerError, "test_email_close_failed")
				}
			}
		} else {
			// 普通连接（无加密）
			err = smtp.SendMail(addr, auth, emailFrom, []string{testEmail}, []byte(message))
		}
	}

	if err != nil {
		errorlog.RecordHTTP(ctx, "config", "Failed to send test email", map[string]any{
			"error": err.Error(),
			"host":  emailHost,
			"port":  emailPort,
			"from":  emailFrom,
			"to":    testEmail,
		}, "Test email error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, "test_email_failed")
	}

	return response.Success(ctx, "test_email_success", http.Json{
		"message": "测试邮件已发送到 " + testEmail,
	})
}
