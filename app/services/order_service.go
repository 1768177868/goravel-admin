package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/oklog/ulid/v2"

	"goravel/app/models"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

type OrderService interface {
	// CreateOrder 创建订单（带防重复提交）
	CreateOrder(userID uint, amount float64, products []OrderProduct, requestID string, remark string) (*models.Order, []models.OrderDetail, error)
	// GetOrderByID 根据ID查询订单
	GetOrderByID(orderID uint, orderTime time.Time) (*models.Order, []models.OrderDetail, error)
	// GetOrderByOrderNo 根据订单号查询订单（直接定位分表，更高效）
	GetOrderByOrderNo(orderNo string) (*models.Order, []models.OrderDetail, error)
	// GetOrders 查询订单列表（限制不超过3个月）
	GetOrders(filters OrderFilters, page, pageSize int) ([]models.Order, int64, error)
	// GetOrdersWithDetails 查询订单列表（包含详情，限制不超过3个月）
	GetOrdersWithDetails(filters OrderFilters, page, pageSize int) ([]OrderWithDetails, int64, error)
	// GetAllOrdersForExport 获取所有订单用于导出（限制不超过3个月，不分页）
	GetAllOrdersForExport(filters OrderFilters) ([]models.Order, error)
	// GetAllOrdersWithDetailsForExport 获取所有订单及详情用于导出（限制不超过3个月，不分页）
	GetAllOrdersWithDetailsForExport(filters OrderFilters) ([]OrderWithDetails, error)
	// UpdateOrderStatus 更新订单状态（已废弃，使用 UpdateOrder）
	UpdateOrderStatus(orderID uint, orderTime time.Time, status string) error
	// UpdateOrder 更新订单（状态和备注）
	UpdateOrder(orderID uint, orderTime time.Time, status string, remark string) error
	// DeleteOrder 删除订单
	DeleteOrder(orderID uint, orderTime time.Time) error
}

// OrderFilters 订单查询筛选条件
type OrderFilters struct {
	UserID    uint      // 用户ID（0表示不筛选）
	OrderNo   string    // 订单号（模糊搜索）
	Status    string    // 订单状态
	MinAmount float64   // 最小金额（0表示不筛选）
	MaxAmount float64   // 最大金额（0表示不筛选）
	StartTime time.Time // 开始时间
	EndTime   time.Time // 结束时间
	OrderBy   string    // 排序字段（格式：字段:asc/desc，如：created_at:desc）
}

type OrderServiceImpl struct {
	shardingService ShardingService
}

type OrderProduct struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

// OrderExportData 订单导出数据结构（用于扩展导出字段）
type OrderExportData struct {
	Order models.Order
}

// OrderWithDetails 订单及详情
type OrderWithDetails struct {
	models.Order
	Details []models.OrderDetail `json:"details"`
}

func NewOrderService() *OrderServiceImpl {
	return &OrderServiceImpl{
		shardingService: NewShardingService(),
	}
}

// CreateOrder 创建订单
func (s *OrderServiceImpl) CreateOrder(userID uint, amount float64, products []OrderProduct, requestID string, remark string) (*models.Order, []models.OrderDetail, error) {
	// 防重复提交：使用 Redis 锁
	if requestID == "" {
		requestID = ulid.Make().String()
	}

	lockKey := fmt.Sprintf("order:lock:%s", requestID)
	lockValue := fmt.Sprintf("%d_%d", userID, time.Now().Unix())

	// 尝试获取锁，过期时间5秒
	// 先检查是否已存在
	var cachedValue string
	cacheResult := facades.Cache().Get(lockKey, &cachedValue)
	if cacheResult != nil && cachedValue != "" {
		return nil, nil, errors.New("订单正在处理中，请勿重复提交")
	}

	// 设置锁，过期时间5秒
	if err := facades.Cache().Put(lockKey, lockValue, 5*time.Second); err != nil {
		errorlog.Record(context.Background(), "order", "获取锁失败", map[string]any{
			"user_id":    userID,
			"request_id": requestID,
			"lock_key":   lockKey,
			"error":      err.Error(),
		}, "获取锁失败: %v", err)
		return nil, nil, fmt.Errorf("获取锁失败: %v", err)
	}

	// 确保释放锁
	defer func() {
		_ = facades.Cache().Forget(lockKey)
	}()

	// 生成订单号并创建订单（带重试机制，防止并发下订单号重复）
	now := time.Now().UTC() // 使用 UTC 时间，确保分表按 UTC 时区分
	tableName := utils.GetShardingTableName("orders", now)

	// 确保分表存在
	if err := s.shardingService.EnsureShardingTable(tableName, "orders"); err != nil {
		_ = facades.Cache().Forget(lockKey)
		return nil, nil, err
	}

	// 重试机制：如果订单号重复，重新生成（最多重试3次）
	var order *models.Order
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		orderNo := s.generateOrderNo()

		// 创建订单主表记录
		order = &models.Order{
			OrderNo: orderNo,
			UserID:  userID,
			Amount:  amount,
			Status:  "pending",
			Remark:  remark,
			// CreatedAt 由 orm.Model 自动设置
		}

		// 尝试插入订单（数据库唯一索引会防止重复）
		err := facades.Orm().Query().Table(tableName).Create(order)
		if err == nil {
			// 插入成功，跳出循环
			break
		}

		// 检查是否是唯一索引冲突错误（订单号重复）
		// MySQL 错误码 1062 表示重复键错误
		errStr := err.Error()
		if strings.Contains(errStr, "Duplicate entry") || strings.Contains(errStr, "1062") {
			// 订单号重复，重试（ULID 碰撞概率极低，但理论上可能发生）
			if i == maxRetries-1 {
				// 最后一次重试也失败，返回错误
				_ = facades.Cache().Forget(lockKey)
				errorlog.Record(context.Background(), "order", "生成唯一订单号失败", map[string]any{
					"user_id":    userID,
					"request_id": requestID,
					"retries":    maxRetries,
				}, "生成唯一订单号失败，请重试")
				return nil, nil, fmt.Errorf("生成唯一订单号失败，请重试")
			}
			// 继续下一次重试
			continue
		}

		// 其他错误，直接返回
		_ = facades.Cache().Forget(lockKey)
		errorlog.Record(context.Background(), "order", "创建订单失败", map[string]any{
			"user_id":    userID,
			"request_id": requestID,
			"amount":     amount,
			"error":      err.Error(),
		}, "创建订单失败: %v", err)
		return nil, nil, fmt.Errorf("创建订单失败: %v", err)
	}

	// 创建订单详情记录
	var details []models.OrderDetail
	detailTableName := utils.GetShardingTableName("order_details", now)

	// 确保订单详情分表存在
	if err := s.shardingService.EnsureShardingTable(detailTableName, "order_details"); err != nil {
		_ = facades.Cache().Forget(lockKey)
		return nil, nil, err
	}

	for _, product := range products {
		detail := models.OrderDetail{
			OrderID:     order.ID,
			ProductID:   product.ProductID,
			ProductName: product.ProductName,
			Price:       product.Price,
			Quantity:    product.Quantity,
			Subtotal:    product.Price * float64(product.Quantity),
			// CreatedAt 由 orm.Model 自动设置
		}

		if err := facades.Orm().Query().Table(detailTableName).Create(&detail); err != nil {
			// 如果详情创建失败，删除已创建的订单
			_, _ = facades.Orm().Query().Table(tableName).Where("id", order.ID).Delete(&models.Order{})
			_ = facades.Cache().Forget(lockKey)
			errorlog.Record(context.Background(), "order", "创建订单详情失败", map[string]any{
				"order_id":   order.ID,
				"user_id":    userID,
				"product_id": product.ProductID,
				"error":      err.Error(),
			}, "创建订单详情失败: %v", err)
			return nil, nil, fmt.Errorf("创建订单详情失败: %v", err)
		}

		details = append(details, detail)
	}

	return order, details, nil
}

// findOrderByID 通过订单ID查找订单
// 如果提供了订单号，会优先使用订单号直接定位分表（更高效）
// 如果没有订单号，则遍历最近几个月的分表
func (s *OrderServiceImpl) findOrderByID(orderID uint, orderNo ...string) (*models.Order, error) {
	// 如果提供了订单号，优先使用订单号直接定位分表（更高效）
	if len(orderNo) > 0 && orderNo[0] != "" {
		order, err := s.findOrderByOrderNo(orderNo[0])
		if err == nil {
			// 如果通过订单号找到了，验证订单ID是否匹配（如果提供了订单ID）
			if orderID > 0 && order.ID != orderID {
				// 订单号找到了但ID不匹配，继续用ID查找
			} else {
				return order, nil
			}
		}
		// 如果通过订单号查找失败，继续用ID查找
	}

	// 如果没有订单号或通过订单号查找失败，使用ID遍历分表
	if orderID == 0 {
		return nil, fmt.Errorf("订单ID不能为空")
	}

	// 查询最近6个月的分表（足够覆盖大部分场景）
	now := time.Now().UTC()
	startTime := now.AddDate(0, -6, 0)
	tableNames := utils.GetShardingTableNames("orders", startTime, now)

	// 从最新的分表开始查询
	for i := len(tableNames) - 1; i >= 0; i-- {
		var order models.Order
		if err := facades.Orm().Query().Table(tableNames[i]).Where("id", orderID).First(&order); err == nil {
			return &order, nil
		}
	}

	return nil, fmt.Errorf("订单不存在")
}

// GetOrderByID 根据ID查询订单
func (s *OrderServiceImpl) GetOrderByID(orderID uint, orderTime time.Time) (*models.Order, []models.OrderDetail, error) {
	// 先查找订单获取 created_at
	order, err := s.findOrderByID(orderID)
	if err != nil {
		return nil, nil, err
	}

	// 使用订单的 created_at 确定分表（将 carbon.DateTime 转换为 time.Time）
	// 通过格式化字符串再解析的方式转换
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := time.Parse("2006-01-02 15:04:05", timeStr)
	utcLoc, _ := time.LoadLocation("UTC")
	createdAt = createdAt.In(utcLoc)

	// 查询订单详情
	detailTableName := utils.GetShardingTableName("order_details", createdAt)
	var details []models.OrderDetail
	if err := facades.Orm().Query().Table(detailTableName).Where("order_id", orderID).Find(&details); err != nil {
		return nil, nil, fmt.Errorf("查询订单详情失败: %v", err)
	}

	return order, details, nil
}

// GetOrderByOrderNo 根据订单号查询订单（直接定位分表，更高效）
func (s *OrderServiceImpl) GetOrderByOrderNo(orderNo string) (*models.Order, []models.OrderDetail, error) {
	// 通过订单号查找订单（直接定位分表）
	order, err := s.findOrderByOrderNo(orderNo)
	if err != nil {
		return nil, nil, err
	}

	// 使用订单的 created_at 确定详情分表（将 carbon.DateTime 转换为 time.Time）
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := time.Parse("2006-01-02 15:04:05", timeStr)
	utcLoc, _ := time.LoadLocation("UTC")
	createdAt = createdAt.In(utcLoc)

	// 查询订单详情
	detailTableName := utils.GetShardingTableName("order_details", createdAt)
	var details []models.OrderDetail
	if err := facades.Orm().Query().Table(detailTableName).Where("order_id", order.ID).Find(&details); err != nil {
		return nil, nil, fmt.Errorf("查询订单详情失败: %v", err)
	}

	return order, details, nil
}

// GetOrders 查询订单列表（限制不超过3个月）
func (s *OrderServiceImpl) GetOrders(filters OrderFilters, page, pageSize int) ([]models.Order, int64, error) {
	// 验证时间范围不超过3个月
	valid, err := utils.ValidateTimeRange(filters.StartTime, filters.EndTime)
	if !valid {
		return nil, 0, err
	}

	// 获取需要查询的所有分表
	tableNames := utils.GetShardingTableNames("orders", filters.StartTime, filters.EndTime)
	if len(tableNames) == 0 {
		return []models.Order{}, 0, nil
	}

	// 如果只有一个分表，直接查询
	if len(tableNames) == 1 {
		return s.querySingleTable(tableNames[0], filters, page, pageSize)
	}

	// 多个分表需要分别查询后合并（简化实现，实际可以优化）
	var allOrders []models.Order
	var total int64

	for _, tableName := range tableNames {
		orders, count, err := s.querySingleTable(tableName, filters, 1, 10000) // 临时查询所有
		if err != nil {
			continue
		}
		allOrders = append(allOrders, orders...)
		total += count
	}

	// 简单分页（实际应该使用更高效的方式）
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(allOrders) {
		return []models.Order{}, total, nil
	}
	if end > len(allOrders) {
		end = len(allOrders)
	}

	return allOrders[start:end], total, nil
}

// buildShardingQuery 构建分表查询条件（辅助函数，减少重复代码）
func (s *OrderServiceImpl) buildShardingQuery(tableName string, filters OrderFilters) orm.Query {
	query := facades.Orm().Query().Table(tableName).
		Where("created_at >= ?", filters.StartTime).
		Where("created_at <= ?", filters.EndTime)

	// 用户ID筛选
	if filters.UserID > 0 {
		query = query.Where("user_id", filters.UserID)
	}

	// 订单号模糊搜索
	if filters.OrderNo != "" {
		query = query.Where("order_no LIKE ?", "%"+filters.OrderNo+"%")
	}

	// 订单状态筛选
	if filters.Status != "" {
		query = query.Where("status", filters.Status)
	}

	// 金额范围筛选
	if filters.MinAmount > 0 {
		query = query.Where("amount >= ?", filters.MinAmount)
	}
	if filters.MaxAmount > 0 {
		query = query.Where("amount <= ?", filters.MaxAmount)
	}

	return query
}

// querySingleTable 查询单个分表
func (s *OrderServiceImpl) querySingleTable(tableName string, filters OrderFilters, page, pageSize int) ([]models.Order, int64, error) {
	// 构建基础查询条件
	query := s.buildShardingQuery(tableName, filters)

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "created_at:desc"
	}
	query = s.applyOrderBy(query, orderBy)

	// 获取总数
	total, err := query.Count()
	if err != nil {
		return nil, 0, err
	}

	// 构建分页查询（重新构建确保使用分表）
	findQuery := s.buildShardingQuery(tableName, filters)
	findQuery = s.applyOrderBy(findQuery, orderBy)

	// 执行分页查询
	var orders []models.Order
	if err := findQuery.Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetOrdersWithDetails 查询订单列表（包含详情，限制不超过3个月）
func (s *OrderServiceImpl) GetOrdersWithDetails(filters OrderFilters, page, pageSize int) ([]OrderWithDetails, int64, error) {
	// 先查询订单列表
	orders, total, err := s.GetOrders(filters, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 批量查询订单详情
	result := make([]OrderWithDetails, len(orders))
	for i, order := range orders {
		// 将订单转换为 OrderWithDetails
		result[i] = OrderWithDetails{
			Order:   order,
			Details: []models.OrderDetail{},
		}

		// 根据订单的 created_at 确定详情分表
		// 通过格式化字符串再解析的方式转换
		timeStr := order.CreatedAt.ToDateTimeString()
		createdAt, _ := time.Parse("2006-01-02 15:04:05", timeStr)
		utcLoc, _ := time.LoadLocation("UTC")
		createdAt = createdAt.In(utcLoc)

		// 查询订单详情
		detailTableName := utils.GetShardingTableName("order_details", createdAt)
		var details []models.OrderDetail
		if err := facades.Orm().Query().Table(detailTableName).Where("order_id", order.ID).Find(&details); err == nil {
			result[i].Details = details
		}
	}

	return result, total, nil
}

// applyOrderBy 应用排序
func (s *OrderServiceImpl) applyOrderBy(query orm.Query, orderBy string) orm.Query {
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
		"order_no":   true,
		"user_id":    true,
		"amount":     true,
		"status":     true,
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

// GetAllOrdersForExport 获取所有订单用于导出（限制不超过3个月，不分页）
func (s *OrderServiceImpl) GetAllOrdersForExport(filters OrderFilters) ([]models.Order, error) {
	// 验证时间范围不超过3个月
	valid, err := utils.ValidateTimeRange(filters.StartTime, filters.EndTime)
	if !valid {
		return nil, err
	}

	// 获取需要查询的所有分表
	tableNames := utils.GetShardingTableNames("orders", filters.StartTime, filters.EndTime)
	if len(tableNames) == 0 {
		return []models.Order{}, nil
	}

	// 查询所有分表并合并结果
	var allOrders []models.Order

	for _, tableName := range tableNames {
		// 使用 buildShardingQuery 构建查询
		query := s.buildShardingQuery(tableName, filters)

		// 应用排序
		orderBy := filters.OrderBy
		if orderBy == "" {
			orderBy = "created_at:desc"
		}
		query = s.applyOrderBy(query, orderBy)

		var orders []models.Order
		if err := query.Find(&orders); err != nil {
			// 如果某个分表查询失败，继续查询其他分表
			continue
		}
		allOrders = append(allOrders, orders...)
	}

	// 按订单时间倒序排序（因为是从多个表合并的）
	// 这里简化处理，如果需要更精确的排序，可以在合并后统一排序

	return allOrders, nil
}

// GetAllOrdersWithDetailsForExport 获取所有订单及详情用于导出（限制不超过3个月，不分页）
func (s *OrderServiceImpl) GetAllOrdersWithDetailsForExport(filters OrderFilters) ([]OrderWithDetails, error) {
	// 验证时间范围不超过3个月
	valid, err := utils.ValidateTimeRange(filters.StartTime, filters.EndTime)
	if !valid {
		return nil, err
	}

	// 获取需要查询的所有分表
	tableNames := utils.GetShardingTableNames("orders", filters.StartTime, filters.EndTime)
	if len(tableNames) == 0 {
		return []OrderWithDetails{}, nil
	}

	// 查询所有分表并合并结果
	var allOrders []models.Order

	for _, tableName := range tableNames {
		// 使用 buildShardingQuery 构建查询
		query := s.buildShardingQuery(tableName, filters)

		// 应用排序
		orderBy := filters.OrderBy
		if orderBy == "" {
			orderBy = "created_at:desc"
		}
		query = s.applyOrderBy(query, orderBy)

		var orders []models.Order
		if err := query.Find(&orders); err != nil {
			// 如果某个分表查询失败，继续查询其他分表
			continue
		}
		allOrders = append(allOrders, orders...)
	}

	// 批量查询订单详情
	result := make([]OrderWithDetails, len(allOrders))
	for i, order := range allOrders {
		result[i] = OrderWithDetails{
			Order:   order,
			Details: []models.OrderDetail{},
		}

		// 根据订单的 created_at 确定详情分表
		// 通过格式化字符串再解析的方式转换
		timeStr := order.CreatedAt.ToDateTimeString()
		createdAt, _ := time.Parse("2006-01-02 15:04:05", timeStr)
		utcLoc, _ := time.LoadLocation("UTC")
		createdAt = createdAt.In(utcLoc)

		// 查询订单详情
		detailTableName := utils.GetShardingTableName("order_details", createdAt)
		var details []models.OrderDetail
		if err := facades.Orm().Query().Table(detailTableName).Where("order_id", order.ID).Find(&details); err == nil {
			result[i].Details = details
		}
	}

	return result, nil
}

// UpdateOrderStatus 更新订单状态（已废弃，保留以兼容旧接口）
func (s *OrderServiceImpl) UpdateOrderStatus(orderID uint, orderTime time.Time, status string) error {
	return s.UpdateOrder(orderID, orderTime, status, "")
}

// UpdateOrder 更新订单（状态和备注）
func (s *OrderServiceImpl) UpdateOrder(orderID uint, orderTime time.Time, status string, remark string) error {
	// 先查找订单获取 created_at
	order, err := s.findOrderByID(orderID)
	if err != nil {
		return err
	}

	// 使用订单的 created_at 确定分表（将 carbon.DateTime 转换为 time.Time）
	// 通过格式化字符串再解析的方式转换
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := time.Parse("2006-01-02 15:04:05", timeStr)
	utcLoc, _ := time.LoadLocation("UTC")
	createdAt = createdAt.In(utcLoc)
	tableName := utils.GetShardingTableName("orders", createdAt)

	// 构建更新数据（始终更新备注，即使为空字符串）
	updateData := map[string]any{
		"status": status,
		"remark": remark,
	}

	_, err = facades.Orm().Query().Table(tableName).
		Where("id", orderID).
		Update(updateData)
	return err
}

// DeleteOrder 删除订单（软删除）
func (s *OrderServiceImpl) DeleteOrder(orderID uint, orderTime time.Time) error {
	// 先查找订单获取 created_at（只查询未删除的订单）
	order, err := s.findOrderByID(orderID)
	if err != nil {
		return err
	}

	// 使用订单的 created_at 确定分表（将 carbon.DateTime 转换为 time.Time）
	// 通过格式化字符串再解析的方式转换
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := time.Parse("2006-01-02 15:04:05", timeStr)
	utcLoc, _ := time.LoadLocation("UTC")
	createdAt = createdAt.In(utcLoc)

	// 软删除订单详情
	detailTableName := utils.GetShardingTableName("order_details", createdAt)
	if _, err := facades.Orm().Query().Table(detailTableName).Where("order_id", orderID).Delete(&models.OrderDetail{}); err != nil {
		errorlog.Record(context.Background(), "order", "删除订单详情失败", map[string]any{
			"order_id": orderID,
			"error":    err.Error(),
		}, "删除订单详情失败: %v", err)
		return fmt.Errorf("删除订单详情失败: %v", err)
	}

	// 软删除订单主表（Goravel 的 Delete 方法会自动使用软删除，如果模型有 SoftDeletes）
	tableName := utils.GetShardingTableName("orders", createdAt)
	_, err = facades.Orm().Query().Table(tableName).Where("id", orderID).Delete(&models.Order{})
	return err
}

// generateOrderNo 生成订单号（格式：ORD + YYYYMM + ULID）
// 例如：ORD20250101ARZ3S0K5M2X9P4Q6R8T1V3W5Y7Z9
// 其中 202501 表示 2025年01月，用于快速定位分表
func (s *OrderServiceImpl) generateOrderNo() string {
	now := time.Now().UTC()
	yearMonth := now.Format("200601") // 格式：202501
	ulidStr := ulid.Make().String()
	return fmt.Sprintf("ORD%s%s", yearMonth, ulidStr)
}

// parseOrderNoYearMonth 从订单号解析年月信息
// 订单号格式：ORD + YYYYMM + ULID
// 返回：年月字符串（如 "202501"）和是否成功解析
func parseOrderNoYearMonth(orderNo string) (string, bool) {
	// 订单号必须以 ORD 开头，且长度至少为 10（ORD + 6位年月 + 至少1位ULID）
	if len(orderNo) < 10 || !strings.HasPrefix(orderNo, "ORD") {
		return "", false
	}

	// 提取年月部分：ORD 后面6位（YYYYMM）
	if len(orderNo) < 9 {
		return "", false
	}
	yearMonth := orderNo[3:9] // ORD(3) + YYYYMM(6)

	// 验证年月格式（简单验证：6位数字）
	if len(yearMonth) != 6 {
		return "", false
	}
	for _, c := range yearMonth {
		if c < '0' || c > '9' {
			return "", false
		}
	}

	return yearMonth, true
}

// findOrderByOrderNo 通过订单号查找订单（直接定位分表）
func (s *OrderServiceImpl) findOrderByOrderNo(orderNo string) (*models.Order, error) {
	// 从订单号解析年月信息
	yearMonth, ok := parseOrderNoYearMonth(orderNo)
	if !ok {
		// 如果无法解析（可能是旧格式订单号），回退到遍历分表的方式
		now := time.Now().UTC()
		startTime := now.AddDate(0, -6, 0)
		tableNames := utils.GetShardingTableNames("orders", startTime, now)

		// 从最新的分表开始查询
		for i := len(tableNames) - 1; i >= 0; i-- {
			var order models.Order
			if err := facades.Orm().Query().Table(tableNames[i]).Where("order_no", orderNo).First(&order); err == nil {
				return &order, nil
			}
		}
		return nil, fmt.Errorf("订单不存在")
	}

	// 解析年月字符串为时间
	// 格式：200601 -> 2025年01月
	parsedTime, err := time.Parse("200601", yearMonth)
	if err != nil {
		// 解析失败，回退到遍历分表
		now := time.Now().UTC()
		startTime := now.AddDate(0, -6, 0)
		tableNames := utils.GetShardingTableNames("orders", startTime, now)

		for i := len(tableNames) - 1; i >= 0; i-- {
			var order models.Order
			if err := facades.Orm().Query().Table(tableNames[i]).Where("order_no", orderNo).First(&order); err == nil {
				return &order, nil
			}
		}
		return nil, fmt.Errorf("订单不存在")
	}

	// 使用解析的年月确定分表
	tableName := utils.GetShardingTableName("orders", parsedTime)

	// 直接查询对应的分表
	var order models.Order
	if err := facades.Orm().Query().Table(tableName).Where("order_no", orderNo).First(&order); err == nil {
		return &order, nil
	}

	return nil, fmt.Errorf("订单不存在")
}
