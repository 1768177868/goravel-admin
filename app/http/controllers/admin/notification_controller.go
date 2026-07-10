package admin

import (
	"regexp"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/spf13/cast"

	apperrors "goravel/app/errors"
	"goravel/app/http/helpers"
	"goravel/app/http/response"
	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/logger"
)

type NotificationController struct {}


func NewNotificationController() *NotificationController {
	return &NotificationController{}
}

func (r *NotificationController) service(ctx http.Context) services.NotificationService {
	return services.NewNotificationServiceImpl(ctx)
}


func (r *NotificationController) Index(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	page := helpers.GetIntQuery(ctx, "page", 1)
	pageSize := helpers.GetIntQuery(ctx, "page_size", 10)
	notifType := ctx.Request().Query("type", "")
	isRead := ctx.Request().Query("is_read", "")
	notifications, total, err := r.service(ctx).List(admin.ID, page, pageSize, notifType, isRead)
	if err != nil {
		logger.ErrorfHTTP(ctx, "list notifications error: %v", err)
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}
	count, err := r.service(ctx).UnreadCount(admin.ID)
	if err != nil {
		logger.ErrorfHTTP(ctx, "unread count error: %v", err)
	}

	return response.Success(ctx, http.Json{
		"notifications": notifications,
		"unread_count":  count,
		"pagination": http.Json{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (r *NotificationController) UnreadCount(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	count, err := r.service(ctx).UnreadCount(admin.ID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}

	return response.Success(ctx, http.Json{
		"count": count,
	})
}

func (r *NotificationController) Recent(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	limit := helpers.GetIntQuery(ctx, "limit", 5)
	notifications, err := r.service(ctx).ListRecent(admin.ID, limit)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrQueryFailed.Code)
	}

	count, _ := r.service(ctx).UnreadCount(admin.ID)

	return response.Success(ctx, http.Json{
		"notifications": notifications,
		"unread_count":  count,
	})
}

func (r *NotificationController) MarkRead(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	id := helpers.GetUintRoute(ctx, "id")
	if id == 0 {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrParamsRequired.Code)
	}

	if err := r.service(ctx).MarkRead(admin.ID, id); err != nil {
		// 使用业务错误类型，直接提取错误码
		if businessErr, ok := apperrors.GetBusinessError(err); ok {
			return response.Error(ctx, http.StatusNotFound, businessErr.Code)
		}
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
	}

	return response.Success(ctx, http.Json{
		"id": id,
	})
}

func (r *NotificationController) MarkAllRead(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	if err := r.service(ctx).MarkAllRead(admin.ID); err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrUpdateFailed.Code)
	}

	return response.Success(ctx)
}

func (r *NotificationController) Store(ctx http.Context) http.Response {
	admin := r.currentAdmin(ctx)
	if admin == nil {
		return response.Error(ctx, http.StatusUnauthorized, apperrors.ErrNotLoggedIn.Code)
	}

	title := ctx.Request().Input("title")
	content := ctx.Request().Input("content")
	notificationType := ctx.Request().Input("type", "announcement")
	if title == "" || content == "" {
		return response.Error(ctx, http.StatusBadRequest, apperrors.ErrParamsRequired.Code)
	}

	// 处理富文本中的图片 URL，将绝对路径转换为相对路径
	// 前端编辑器可能插入了 http://localhost:3008/api/admin/public/images/14
	// 我们希望数据库存的是 /api/admin/public/images/14
	// 这样即使域名变更，只要相对路径不变，就能正常显示（前端会自动拼接域名）
	// 或者，更进一步，我们也可以存完整路径，但通常存相对路径更灵活。
	// 不过，由于前端展示时，如果已经是 http 开头，前端不会再拼接。
	// 如果我们存相对路径，前端展示时必须知道要拼接域名。
	// 当前 WangEditor.vue 的逻辑是：如果 url 不是 http 开头，则拼接 BASE_URL。
	// 所以，我们可以安全地将 localhost 域名去除，只保留 /api/...

	// 获取当前 APP_URL (虽然这里可能不准确，主要还是去除请求中的域名部分)
	// 简单做法：正则替换 http(s)://[^/]+
	// 但要注意，如果引用了外部图片（如 OSS），不应该替换。
	// 只替换指向本站的图片。

	// 1. 使用正则替换明确的图片接口路径，这是最可靠的方式
	// 匹配 http://.../api/admin/public/images/ 并替换为 /api/admin/public/images/
	re := regexp.MustCompile(`https?://[^/]+(/api/admin/public/images/)`)
	content = re.ReplaceAllString(content, "$1")

	appURL := facades.Config().GetString("app.url")
	if appURL != "" {
		appURL = strings.TrimSuffix(appURL, "/")
		content = strings.ReplaceAll(content, appURL, "")
	}
	// 同时也尝试替换 localhost 形式，以防开发环境
	// 这里比较 hack，更严谨的做法是前端提交时处理，或者后端解析 HTML
	// 简单起见，我们假设只替换当前域名的引用
	host := ctx.Request().Header("Host", "")
	if host != "" {
		scheme := "http"
		if ctx.Request().Header("X-Forwarded-Proto", "") == "https" {
			scheme = "https"
		}
		// 替换 localhost:3008 和 localhost:3007
		// 注意：如果图片链接是 http://localhost:3008/api/... 而访问是 http://localhost:3007
		// 替换时要小心。
		// 这里我们只替换指向后端服务的 URL 前缀

		// 1. 替换 http(s)://Host/
		currentBaseURL := scheme + "://" + host + "/"
		content = strings.ReplaceAll(content, currentBaseURL, "/")

		// 2. 替换 http(s)://Host (无结尾斜杠)
		currentBaseURLNoSlash := scheme + "://" + host
		content = strings.ReplaceAll(content, currentBaseURLNoSlash, "")
	}

	var receiverID *uint
	receiverVal := ctx.Request().Input("receiver_id")
	if receiverVal != "" {
		id := cast.ToUint(receiverVal)
		if id > 0 {
			receiverID = &id
		}
	}

	senderID := admin.ID
	notification, err := r.service(ctx).Create(title, content, notificationType, &senderID, receiverID)
	if err != nil {
		return response.Error(ctx, http.StatusInternalServerError, apperrors.ErrCreateFailed.Code)
	}

	if notification == nil {
		return response.Success(ctx)
	}

	return response.Success(ctx, http.Json{
		"notification": notification,
	})
}

func (r *NotificationController) currentAdmin(ctx http.Context) *models.Admin {
	if adminValue := ctx.Value("admin"); adminValue != nil {
		if admin, ok := adminValue.(models.Admin); ok {
			return &admin
		}
		if adminPtr, ok := adminValue.(*models.Admin); ok {
			return adminPtr
		}
	}
	return nil
}
