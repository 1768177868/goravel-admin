package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/oklog/ulid/v2"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/utils/errorlog"
)

type PaymentService interface {
	// GetPaymentMethodByID 根据ID获取支付方式
	GetPaymentMethodByID(id uint) (*models.PaymentMethod, error)
	// GetPaymentMethodByCode 根据代码获取支付方式
	GetPaymentMethodByCode(code string) (*models.PaymentMethod, error)
	// GetPaymentMethods 获取支付方式列表
	GetPaymentMethods(filters PaymentMethodFilters, page, pageSize int) ([]models.PaymentMethod, int64, error)
	// CreatePaymentMethod 创建支付方式
	CreatePaymentMethod(name, code, paymentType string, config map[string]any, isActive bool, sort int, description string) (*models.PaymentMethod, error)
	// UpdatePaymentMethod 更新支付方式
	UpdatePaymentMethod(id uint, name string, config map[string]any, isActive bool, sort int, description string) error
	// DeletePaymentMethod 删除支付方式
	DeletePaymentMethod(id uint) error

	// GetPaymentByID 根据ID获取支付记录
	GetPaymentByID(id uint) (*models.Payment, error)
	// GetPaymentByPaymentNo 根据支付单号获取支付记录
	GetPaymentByPaymentNo(paymentNo string) (*models.Payment, error)
	// GetPayments 获取支付记录列表
	GetPayments(filters PaymentFilters, page, pageSize int) ([]models.Payment, int64, error)
	// CreatePayment 创建支付记录
	CreatePayment(orderNo string, paymentMethodID uint, userID uint, amount float64, remark string) (*models.Payment, error)
	// UpdatePaymentStatus 更新支付状态
	UpdatePaymentStatus(paymentID uint, status string, thirdPartyNo string, payTime *time.Time, failReason string, notifyData map[string]any) error

	// CreatePaymentOrder 创建支付订单（调用第三方支付）
	CreatePaymentOrder(payment *models.Payment, clientIP string) (map[string]any, error)
	// QueryPaymentOrder 查询支付订单状态
	QueryPaymentOrder(payment *models.Payment) (map[string]any, error)
	// HandlePaymentNotify 处理支付回调通知
	HandlePaymentNotify(paymentMethod *models.PaymentMethod, notifyData map[string]any) (*models.Payment, error)
}

// PaymentMethodFilters 支付方式查询过滤器
type PaymentMethodFilters struct {
	Name        string
	Code        string
	Type        string
	IsActive    string
	Description string
	OrderBy     string
}

// PaymentFilters 支付记录查询过滤器
type PaymentFilters struct {
	PaymentNo       string
	OrderNo         string
	PaymentMethodID uint
	UserID          uint
	Status          string
	StartTime       time.Time
	EndTime         time.Time
	OrderBy         string
}

type PaymentServiceImpl struct{}

func NewPaymentService() PaymentService {
	return &PaymentServiceImpl{}
}

// GetPaymentMethodByID 根据ID获取支付方式
func (s *PaymentServiceImpl) GetPaymentMethodByID(id uint) (*models.PaymentMethod, error) {
	var paymentMethod models.PaymentMethod
	if err := facades.Orm().Query().Where("id", id).First(&paymentMethod); err != nil {
		return nil, apperrors.ErrPaymentMethodNotFound.WithError(err)
	}
	return &paymentMethod, nil
}

// GetPaymentMethodByCode 根据代码获取支付方式
func (s *PaymentServiceImpl) GetPaymentMethodByCode(code string) (*models.PaymentMethod, error) {
	var paymentMethod models.PaymentMethod
	if err := facades.Orm().Query().Where("code", code).Where("is_active", true).First(&paymentMethod); err != nil {
		return nil, apperrors.ErrPaymentMethodNotFound.WithError(err)
	}
	return &paymentMethod, nil
}

// GetPaymentMethods 获取支付方式列表
func (s *PaymentServiceImpl) GetPaymentMethods(filters PaymentMethodFilters, page, pageSize int) ([]models.PaymentMethod, int64, error) {
	query := facades.Orm().Query().Model(&models.PaymentMethod{})

	// 应用筛选条件
	if filters.Name != "" {
		query = query.Where("name LIKE ?", "%"+filters.Name+"%")
	}
	if filters.Code != "" {
		query = query.Where("code", filters.Code)
	}
	if filters.Type != "" {
		query = query.Where("type", filters.Type)
	}
	if filters.IsActive != "" {
		if filters.IsActive == "1" {
			query = query.Where("is_active", true)
		} else if filters.IsActive == "0" {
			query = query.Where("is_active", false)
		}
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, apperrors.ErrQueryFailed.WithError(err)
	}

	// 应用排序
	if filters.OrderBy != "" {
		query = s.applyOrderBy(query, filters.OrderBy)
	} else {
		query = query.Order("sort asc").Order("id desc")
	}

	// 分页查询
	var paymentMethods []models.PaymentMethod
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&paymentMethods); err != nil {
		return nil, 0, apperrors.ErrQueryFailed.WithError(err)
	}

	return paymentMethods, total, nil
}

// CreatePaymentMethod 创建支付方式
func (s *PaymentServiceImpl) CreatePaymentMethod(name, code, paymentType string, config map[string]any, isActive bool, sort int, description string) (*models.PaymentMethod, error) {
	// 检查代码是否已存在
	var existing models.PaymentMethod
	if err := facades.Orm().Query().Where("code", code).First(&existing); err == nil {
		return nil, apperrors.ErrPaymentMethodCodeExists
	}

	// 序列化配置
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, apperrors.ErrPaymentConfigRequired.WithError(err)
	}

	paymentMethod := &models.PaymentMethod{
		Name:        name,
		Code:        code,
		Type:        paymentType,
		Config:      string(configJSON),
		IsActive:    isActive,
		Sort:        sort,
		Description: description,
	}

	if err := facades.Orm().Query().Create(paymentMethod); err != nil {
		return nil, apperrors.ErrCreateFailed.WithError(err)
	}

	return paymentMethod, nil
}

// UpdatePaymentMethod 更新支付方式
func (s *PaymentServiceImpl) UpdatePaymentMethod(id uint, name string, config map[string]any, isActive bool, sort int, description string) error {
	paymentMethod, err := s.GetPaymentMethodByID(id)
	if err != nil {
		return err
	}

	// 序列化配置
	var configJSON string
	if config != nil {
		configBytes, err := json.Marshal(config)
		if err != nil {
			return apperrors.ErrPaymentConfigRequired.WithError(err)
		}
		configJSON = string(configBytes)
	} else {
		configJSON = paymentMethod.Config // 保持原有配置
	}

	updateData := map[string]any{
		"name":        name,
		"config":      configJSON,
		"is_active":   isActive,
		"sort":        sort,
		"description": description,
	}

	if _, err := facades.Orm().Query().Where("id", id).Update(&models.PaymentMethod{}, updateData); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}

	return nil
}

// DeletePaymentMethod 删除支付方式
func (s *PaymentServiceImpl) DeletePaymentMethod(id uint) error {
	_, err := s.GetPaymentMethodByID(id)
	if err != nil {
		return err
	}

	if _, err := facades.Orm().Query().Where("id", id).Delete(&models.PaymentMethod{}); err != nil {
		return apperrors.ErrDeleteFailed.WithError(err)
	}

	return nil
}

// GetPaymentByID 根据ID获取支付记录
func (s *PaymentServiceImpl) GetPaymentByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	if err := facades.Orm().Query().Where("id", id).First(&payment); err != nil {
		return nil, apperrors.ErrPaymentNotFound.WithError(err)
	}
	// 手动加载支付方式
	if payment.PaymentMethodID > 0 {
		paymentMethod, err := s.GetPaymentMethodByID(payment.PaymentMethodID)
		if err == nil {
			payment.PaymentMethod = *paymentMethod
		}
	}
	return &payment, nil
}

// GetPaymentByPaymentNo 根据支付单号获取支付记录
func (s *PaymentServiceImpl) GetPaymentByPaymentNo(paymentNo string) (*models.Payment, error) {
	var payment models.Payment
	if err := facades.Orm().Query().Where("payment_no", paymentNo).First(&payment); err != nil {
		return nil, apperrors.ErrPaymentNotFound.WithError(err)
	}
	// 手动加载支付方式
	if payment.PaymentMethodID > 0 {
		paymentMethod, err := s.GetPaymentMethodByID(payment.PaymentMethodID)
		if err == nil {
			payment.PaymentMethod = *paymentMethod
		}
	}
	return &payment, nil
}

// GetPayments 获取支付记录列表
func (s *PaymentServiceImpl) GetPayments(filters PaymentFilters, page, pageSize int) ([]models.Payment, int64, error) {
	query := facades.Orm().Query().Model(&models.Payment{})

	// 应用筛选条件
	if filters.PaymentNo != "" {
		query = query.Where("payment_no LIKE ?", "%"+filters.PaymentNo+"%")
	}
	if filters.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+filters.OrderNo+"%")
	}
	if filters.PaymentMethodID > 0 {
		query = query.Where("payment_method_id", filters.PaymentMethodID)
	}
	if filters.UserID > 0 {
		query = query.Where("user_id", filters.UserID)
	}
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
	}
	if !filters.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filters.EndTime)
	}

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, apperrors.ErrQueryFailed.WithError(err)
	}

	// 应用排序
	if filters.OrderBy != "" {
		query = s.applyOrderBy(query, filters.OrderBy)
	} else {
		query = query.Order("created_at desc")
	}

	// 分页查询
	var payments []models.Payment
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&payments); err != nil {
		return nil, 0, apperrors.ErrQueryFailed.WithError(err)
	}

	// 批量加载支付方式
	paymentMethodIDs := make(map[uint]bool)
	for _, payment := range payments {
		if payment.PaymentMethodID > 0 {
			paymentMethodIDs[payment.PaymentMethodID] = true
		}
	}

	// 批量查询支付方式
	paymentMethodsMap := make(map[uint]*models.PaymentMethod)
	if len(paymentMethodIDs) > 0 {
		var ids []uint
		for id := range paymentMethodIDs {
			ids = append(ids, id)
		}
		// 转换为 []any
		idsAny := make([]any, len(ids))
		for i, id := range ids {
			idsAny[i] = id
		}
		var paymentMethods []models.PaymentMethod
		if err := facades.Orm().Query().Model(&models.PaymentMethod{}).WhereIn("id", idsAny).Find(&paymentMethods); err == nil {
			for i := range paymentMethods {
				paymentMethodsMap[paymentMethods[i].ID] = &paymentMethods[i]
			}
		}
	}

	// 关联支付方式
	for i := range payments {
		if pm, ok := paymentMethodsMap[payments[i].PaymentMethodID]; ok {
			payments[i].PaymentMethod = *pm
		}
	}

	return payments, total, nil
}

// CreatePayment 创建支付记录
func (s *PaymentServiceImpl) CreatePayment(orderNo string, paymentMethodID uint, userID uint, amount float64, remark string) (*models.Payment, error) {
	// 验证金额
	if amount <= 0 {
		return nil, apperrors.ErrPaymentAmountInvalid
	}

	// 验证支付方式
	paymentMethod, err := s.GetPaymentMethodByID(paymentMethodID)
	if err != nil {
		return nil, err
	}
	if !paymentMethod.IsActive {
		return nil, apperrors.ErrPaymentMethodDisabled
	}

	// 生成支付单号
	paymentNo := fmt.Sprintf("PAY%s%s", time.Now().Format("20060102"), ulid.Make().String())

	payment := &models.Payment{
		PaymentNo:       paymentNo,
		OrderNo:         orderNo,
		PaymentMethodID: paymentMethodID,
		UserID:          userID,
		Amount:          amount,
		Status:          "pending",
		Remark:          remark,
	}

	if err := facades.Orm().Query().Create(payment); err != nil {
		errorlog.Record(context.Background(), "payment", "创建支付记录失败", map[string]any{
			"order_no":          orderNo,
			"payment_method_id": paymentMethodID,
			"user_id":           userID,
			"amount":            amount,
			"error":             err.Error(),
		}, "创建支付记录失败: %v", err)
		return nil, apperrors.ErrCreatePaymentFailed.WithError(err)
	}

	return payment, nil
}

// UpdatePaymentStatus 更新支付状态
func (s *PaymentServiceImpl) UpdatePaymentStatus(paymentID uint, status string, thirdPartyNo string, payTime *time.Time, failReason string, notifyData map[string]any) error {
	updateData := map[string]any{
		"status": status,
	}

	if thirdPartyNo != "" {
		updateData["third_party_no"] = thirdPartyNo
	}
	if payTime != nil {
		updateData["pay_time"] = payTime
	}
	if failReason != "" {
		updateData["fail_reason"] = failReason
	}
	if notifyData != nil {
		notifyJSON, err := json.Marshal(notifyData)
		if err == nil {
			updateData["notify_data"] = string(notifyJSON)
		}
	}

	if _, err := facades.Orm().Query().Where("id", paymentID).Update(&models.Payment{}, updateData); err != nil {
		return apperrors.ErrUpdateFailed.WithError(err)
	}

	return nil
}

// CreatePaymentOrder 创建支付订单（调用第三方支付）
func (s *PaymentServiceImpl) CreatePaymentOrder(payment *models.Payment, clientIP string) (map[string]any, error) {
	// 获取支付方式
	paymentMethod, err := s.GetPaymentMethodByID(payment.PaymentMethodID)
	if err != nil {
		return nil, err
	}

	// 解析配置
	var config map[string]any
	if err := json.Unmarshal([]byte(paymentMethod.Config), &config); err != nil {
		return nil, apperrors.ErrPaymentConfigRequired.WithError(err)
	}

	// 根据支付类型调用不同的支付接口
	switch paymentMethod.Type {
	case "wechat":
		return s.createWechatPayment(payment, config, clientIP)
	case "alipay":
		return s.createAlipayPayment(payment, config, clientIP)
	default:
		return nil, apperrors.ErrInvalidPaymentType.WithMessage(fmt.Sprintf("不支持的支付类型: %s", paymentMethod.Type))
	}
}

// createWechatPayment 创建微信支付订单
func (s *PaymentServiceImpl) createWechatPayment(payment *models.Payment, config map[string]any, clientIP string) (map[string]any, error) {
	// 这里需要根据 gopay 的微信支付文档实现
	// 示例代码，实际使用时需要根据 gopay 的最新 API 调整
	// 参考: https://github.com/go-pay/gopay/tree/main/wechat/v3

	// 获取配置参数
	appID, _ := config["app_id"].(string)
	mchID, _ := config["mch_id"].(string)
	apiV3Key, _ := config["api_v3_key"].(string)
	certSerialNo, _ := config["cert_serial_no"].(string)
	privateKeyPath, _ := config["private_key_path"].(string)

	if appID == "" || mchID == "" || apiV3Key == "" {
		return nil, apperrors.ErrPaymentConfigRequired.WithMessage("微信支付配置不完整")
	}

	// 创建微信支付客户端
	client, err := wechat.NewClientV3(appID, mchID, apiV3Key, certSerialNo)
	if err != nil {
		return nil, apperrors.ErrCreatePaymentFailed.WithError(err)
	}
	// 设置私钥（如果需要）
	if privateKeyPath != "" {
		// 这里需要根据 gopay 的实际 API 设置私钥
		// client.SetPrivateKey(...)
	}

	// 设置回调地址（需要从配置中读取）
	notifyURL, _ := config["notify_url"].(string)
	if notifyURL == "" {
		notifyURL = fmt.Sprintf("%s/api/payment/notify/wechat", facades.Config().GetString("app.url"))
	}

	// 创建支付订单
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", payment.PaymentNo)
	bm.Set("description", payment.Remark)
	bm.Set("amount", map[string]any{
		"total":    int(payment.Amount * 100), // 转换为分
		"currency": "CNY",
	})
	bm.Set("notify_url", notifyURL)
	bm.Set("payer", map[string]any{
		"openid": config["openid"], // 需要从订单或用户信息中获取
	})

	// 注意：这里需要根据 gopay 的最新 API 调用
	// 示例代码，实际使用时需要根据 gopay 的最新文档调整
	ctx := context.Background()
	wxRsp, err := client.V3TransactionJsapi(ctx, bm)
	if err != nil {
		return nil, apperrors.ErrCreatePaymentFailed.WithError(err)
	}

	if wxRsp.Code != wechat.Success {
		return nil, apperrors.ErrCreatePaymentFailed.WithMessage(wxRsp.Error)
	}

	return map[string]any{
		"payment_no": payment.PaymentNo,
		"prepay_id":  wxRsp.Response.PrepayId,
		// PaySign 可能需要从其他地方获取或计算
	}, nil
}

// createAlipayPayment 创建支付宝支付订单
func (s *PaymentServiceImpl) createAlipayPayment(payment *models.Payment, config map[string]any, clientIP string) (map[string]any, error) {
	// 这里需要根据 gopay 的支付宝文档实现
	// 示例代码，实际使用时需要根据 gopay 的最新 API 调整
	// 参考: https://github.com/go-pay/gopay/tree/main/alipay

	// 获取配置参数
	appID, _ := config["app_id"].(string)
	privateKey, _ := config["private_key"].(string)
	// appCertPublicKey, _ := config["app_cert_public_key"].(string)
	// alipayRootCert, _ := config["alipay_root_cert"].(string)
	// alipayPublicCert, _ := config["alipay_public_cert"].(string)

	if appID == "" || privateKey == "" {
		return nil, apperrors.ErrPaymentConfigRequired.WithMessage("支付宝配置不完整")
	}

	// 创建支付宝客户端
	client, err := alipay.NewClient(appID, privateKey, false)
	if err != nil {
		return nil, apperrors.ErrCreatePaymentFailed.WithError(err)
	}

	// 设置证书（如果需要）
	// 注意：这里需要根据 gopay 的最新 API 设置证书
	// 示例代码，实际使用时需要根据 gopay 的最新文档调整
	// if appCertPublicKey != "" {
	// 	client.SetAppCertPublicKey(appCertPublicKey)
	// }
	// if alipayRootCert != "" {
	// 	client.SetAlipayRootCert(alipayRootCert)
	// }
	// if alipayPublicCert != "" {
	// 	client.SetAlipayPublicCert(alipayPublicCert)
	// }

	// 设置回调地址
	notifyURL, _ := config["notify_url"].(string)
	if notifyURL == "" {
		notifyURL = fmt.Sprintf("%s/api/payment/notify/alipay", facades.Config().GetString("app.url"))
	}
	client.SetNotifyUrl(notifyURL)

	// 创建支付订单
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", payment.PaymentNo)
	bm.Set("subject", payment.Remark)
	bm.Set("total_amount", fmt.Sprintf("%.2f", payment.Amount))
	bm.Set("product_code", "QUICK_MSECURITY_PAY")

	ctx := context.Background()
	payUrl, err := client.TradeAppPay(ctx, bm)
	if err != nil {
		return nil, apperrors.ErrCreatePaymentFailed.WithError(err)
	}

	return map[string]any{
		"payment_no": payment.PaymentNo,
		"pay_url":    payUrl,
	}, nil
}

// QueryPaymentOrder 查询支付订单状态
func (s *PaymentServiceImpl) QueryPaymentOrder(payment *models.Payment) (map[string]any, error) {
	// 获取支付方式
	paymentMethod, err := s.GetPaymentMethodByID(payment.PaymentMethodID)
	if err != nil {
		return nil, err
	}

	// 解析配置
	var config map[string]any
	if err := json.Unmarshal([]byte(paymentMethod.Config), &config); err != nil {
		return nil, apperrors.ErrPaymentConfigRequired.WithError(err)
	}

	// 根据支付类型查询
	switch paymentMethod.Type {
	case "wechat":
		return s.queryWechatPayment(payment, config)
	case "alipay":
		return s.queryAlipayPayment(payment, config)
	default:
		return nil, apperrors.ErrInvalidPaymentType.WithMessage(fmt.Sprintf("不支持的支付类型: %s", paymentMethod.Type))
	}
}

// queryWechatPayment 查询微信支付订单状态
func (s *PaymentServiceImpl) queryWechatPayment(payment *models.Payment, config map[string]any) (map[string]any, error) {
	// 实现微信支付查询逻辑
	// 这里需要根据 gopay 的微信支付文档实现
	return nil, fmt.Errorf("微信支付查询功能待实现")
}

// queryAlipayPayment 查询支付宝支付订单状态
func (s *PaymentServiceImpl) queryAlipayPayment(payment *models.Payment, config map[string]any) (map[string]any, error) {
	// 实现支付宝支付查询逻辑
	// 这里需要根据 gopay 的支付宝文档实现
	return nil, fmt.Errorf("支付宝支付查询功能待实现")
}

// HandlePaymentNotify 处理支付回调通知
func (s *PaymentServiceImpl) HandlePaymentNotify(paymentMethod *models.PaymentMethod, notifyData map[string]any) (*models.Payment, error) {
	// 根据支付类型处理回调
	switch paymentMethod.Type {
	case "wechat":
		return s.handleWechatNotify(paymentMethod, notifyData)
	case "alipay":
		return s.handleAlipayNotify(paymentMethod, notifyData)
	default:
		return nil, apperrors.ErrInvalidPaymentType.WithMessage(fmt.Sprintf("不支持的支付类型: %s", paymentMethod.Type))
	}
}

// handleWechatNotify 处理微信支付回调
func (s *PaymentServiceImpl) handleWechatNotify(paymentMethod *models.PaymentMethod, notifyData map[string]any) (*models.Payment, error) {
	// 实现微信支付回调处理逻辑
	// 这里需要根据 gopay 的微信支付文档实现
	return nil, fmt.Errorf("微信支付回调处理功能待实现")
}

// handleAlipayNotify 处理支付宝支付回调
func (s *PaymentServiceImpl) handleAlipayNotify(paymentMethod *models.PaymentMethod, notifyData map[string]any) (*models.Payment, error) {
	// 实现支付宝支付回调处理逻辑
	// 这里需要根据 gopay 的支付宝文档实现
	return nil, fmt.Errorf("支付宝支付回调处理功能待实现")
}

// applyOrderBy 应用排序
func (s *PaymentServiceImpl) applyOrderBy(query orm.Query, orderBy string) orm.Query {
	// 解析排序字段，格式：字段:asc/desc
	parts := strings.Split(orderBy, ":")
	if len(parts) != 2 {
		// 默认排序
		return query.Order("created_at desc")
	}

	field := parts[0]
	direction := strings.ToLower(parts[1])

	// 允许排序的字段
	allowedFields := map[string]bool{
		"id":         true,
		"name":       true,
		"code":       true,
		"type":       true,
		"is_active":  true,
		"sort":       true,
		"created_at": true,
		"updated_at": true,
	}

	if !allowedFields[field] {
		// 如果字段不允许，使用默认排序
		return query.Order("created_at desc")
	}

	if direction == "asc" {
		return query.Order(field + " asc")
	} else {
		return query.Order(field + " desc")
	}
}
