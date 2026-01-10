package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/oklog/ulid/v2"

	apperrors "goravel/app/errors"
	"goravel/app/models"
	"goravel/app/utils"
	"goravel/app/utils/errorlog"
)

// ApplyOrderFiltersToQuery 只负责通用筛选（不包含时间范围），供列表查询/导出复用，避免重复/不一致。
func ApplyOrderFiltersToQuery(query orm.Query, filters OrderFilters) orm.Query {
	// 用户ID筛选
	if filters.UserID > 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}

	// 订单号模糊搜索
	if filters.OrderNo != "" {
		query = query.Where("order_no = ?", filters.OrderNo)
	}

	// 订单状态筛选
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
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

// BuildOrderQuery 构建订单分表查询（包含时间范围 + 通用筛选），供列表查询/导出复用。
func BuildOrderQuery(tableName string, filters OrderFilters) orm.Query {
	query := facades.Orm().Query().Table(tableName)

	// 时间范围（导出/列表都需要）
	if !filters.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filters.StartTime)
	}
	if !filters.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filters.EndTime)
	}

	// 商品名称搜索（通过子查询在 order_details 表中查找）
	if filters.ProductName != "" {
		// 获取对应的 order_details 分表名
		detailTableName := GetOrderDetailsTableFromOrdersTable(tableName)
		query = query.Where("id IN (SELECT order_id FROM "+detailTableName+" WHERE product_name = ?)", filters.ProductName)
	}

	return ApplyOrderFiltersToQuery(query, filters)
}

// GetOrderDetailsTableFromOrdersTable 从订单分表名获取对应的订单详情分表名
// ordersTableName: 订单分表名，如 "orders_202501"
// 返回: 订单详情分表名，如 "order_details_202501"
func GetOrderDetailsTableFromOrdersTable(ordersTableName string) string {
	// 将 orders_YYYYMM 转换为 order_details_YYYYMM
	return strings.Replace(ordersTableName, "orders_", "order_details_", 1)
}

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
	// 如果提供了订单号，优先使用订单号查找订单（更高效，可直接定位分表）
	UpdateOrder(orderID uint, orderTime time.Time, status string, remark string, orderNo ...string) error
	// DeleteOrder 删除订单
	// 如果提供了订单号，优先使用订单号查找订单（更高效，可直接定位分表）
	DeleteOrder(orderID uint, orderTime time.Time, orderNo ...string) error
	// GetOrdersCountInYear 获取最近一年的订单总数（用于仪表盘统计）
	GetOrdersCountInYear() (int64, error)
}

// OrderFilters 订单查询筛选条件
type OrderFilters struct {
	UserID      uint      // 用户ID（0表示不筛选）
	OrderNo     string    // 订单号（模糊搜索）
	ProductName string    // 商品名称（模糊搜索）
	Status      string    // 订单状态
	MinAmount   float64   // 最小金额（0表示不筛选）
	MaxAmount   float64   // 最大金额（0表示不筛选）
	StartTime   time.Time // 开始时间
	EndTime     time.Time // 结束时间
	OrderBy     string    // 排序字段（格式：字段:asc/desc，如：created_at:desc）
}

// 订单分页统计优化阈值（超过此值使用执行计划估算）
const OrderCountThreshold int64 = 100000

type OrderServiceImpl struct {
	shardingService      ShardingService
	shardingQueryService ShardingQueryService
	countThreshold       int64 // count 查询优化阈值
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
	service := &OrderServiceImpl{
		shardingService: NewShardingService(),
	}

	// 初始化分表查询服务
	service.shardingQueryService = NewShardingQueryService(ShardingQueryConfig{
		BaseTableName: "orders",
		GetColumns: func() string {
			return service.getOrderTableColumns()
		},
		BuildWhereClause: func(filters any) (string, []any) {
			return service.buildOrderWhereClause(filters)
		},
		GetAllowedOrderFields: func() map[string]bool {
			return map[string]bool{
				"id":         true,
				"order_no":   true,
				"user_id":    true,
				"amount":     true,
				"status":     true,
				"created_at": true,
				"updated_at": true,
			}
		},
		DefaultOrderBy: "created_at:desc",
		ModuleName:     "order",
		CountThreshold: OrderCountThreshold, // 订单数据量大，超过此值使用估算值，不设置或者0直接使用count
	})

	return service
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
		return nil, nil, apperrors.ErrGetLockFailed.WithError(err)
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
	for i := range maxRetries {
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
				return nil, nil, apperrors.ErrGenerateOrderNoFailed
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
		return nil, nil, apperrors.ErrCreateOrderFailed.WithError(err)
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
			return nil, nil, apperrors.ErrCreateOrderDetailFailed.WithError(err)
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
		return nil, apperrors.ErrOrderIDRequired
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

	return nil, apperrors.ErrOrderNotFound
}

// GetOrderByID 根据ID查询订单
func (s *OrderServiceImpl) GetOrderByID(orderID uint, orderTime time.Time) (*models.Order, []models.OrderDetail, error) {
	// 先查找订单获取 created_at
	order, err := s.findOrderByID(orderID)
	if err != nil {
		return nil, nil, err
	}

	// 使用订单的 created_at 确定分表
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := utils.ParseDateTimeUTC(timeStr)

	// 查询订单详情
	detailTableName := utils.GetShardingTableName("order_details", createdAt)
	var details []models.OrderDetail
	if err := facades.Orm().Query().Table(detailTableName).Where("order_id", orderID).Find(&details); err != nil {
		return nil, nil, apperrors.ErrQueryOrderDetailFailed.WithError(err)
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

	// 使用订单的 created_at 确定详情分表
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := utils.ParseDateTimeUTC(timeStr)

	// 查询订单详情
	detailTableName := utils.GetShardingTableName("order_details", createdAt)
	var details []models.OrderDetail
	if err := facades.Orm().Query().Table(detailTableName).Where("order_id", order.ID).Find(&details); err != nil {
		return nil, nil, apperrors.ErrQueryOrderDetailFailed.WithError(err)
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

	// 如果有商品名搜索，使用单表逐个查询方式（因为需要关联不同的 order_details 分表）
	if filters.ProductName != "" {
		return s.queryMultipleTablesSequentially(tableNames, filters, page, pageSize)
	}

	// 多个分表：使用 UNION ALL 在数据库层面合并
	return s.queryMultipleTablesWithUnion(tableNames, filters, page, pageSize)
}

// buildShardingQuery 构建分表查询条件（辅助函数，减少重复代码）
func (s *OrderServiceImpl) buildShardingQuery(tableName string, filters OrderFilters) orm.Query {
	return BuildOrderQuery(tableName, filters)
}

// buildOrderWhereClause 构建订单查询的 WHERE 条件（用于通用分表查询服务）
func (s *OrderServiceImpl) buildOrderWhereClause(filters any) (string, []any) {
	orderFilters, ok := filters.(OrderFilters)
	if !ok {
		return "", nil
	}

	var conditions []string
	var args []any

	// 时间范围（必填）
	if !orderFilters.StartTime.IsZero() {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, orderFilters.StartTime)
	}
	if !orderFilters.EndTime.IsZero() {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, orderFilters.EndTime)
	}

	// 用户ID筛选
	if orderFilters.UserID > 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, orderFilters.UserID)
	}

	// 订单号模糊搜索
	if orderFilters.OrderNo != "" {
		// conditions = append(conditions, "order_no LIKE ?")
		// args = append(args, "%"+orderFilters.OrderNo+"%")
		conditions = append(conditions, "order_no = ?")
		args = append(args, orderFilters.OrderNo)
	}

	// 订单状态筛选
	if orderFilters.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, orderFilters.Status)
	}

	// 金额范围筛选
	if orderFilters.MinAmount > 0 {
		conditions = append(conditions, "amount >= ?")
		args = append(args, orderFilters.MinAmount)
	}
	if orderFilters.MaxAmount > 0 {
		conditions = append(conditions, "amount <= ?")
		args = append(args, orderFilters.MaxAmount)
	}

	return strings.Join(conditions, " AND "), args
}

// getOrderTableColumns 获取订单表的所有列名（用于 UNION ALL 查询）
// 明确指定列名，避免不同分表列数不一致的问题
// 只查询 Order 模型中存在的字段，忽略可能不存在的扩展字段（如 payment_method）
func (s *OrderServiceImpl) getOrderTableColumns() string {
	// 列顺序必须与 Order 模型字段顺序一致
	// 只包含 Order 模型中定义的字段，确保所有分表都有这些字段
	// 注意：不包含 payment_method，因为：
	// 1. Order 模型中没有该字段
	// 2. 某些旧分表可能没有该字段
	// 3. 如果将来需要该字段，应该先更新 Order 模型，然后确保所有分表都有该字段
	columns := []string{
		"id",
		"order_no",
		"user_id",
		"amount",
		"status",
		"remark",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	return strings.Join(columns, ", ")
}

// queryMultipleTablesWithUnion 使用 UNION ALL 查询多个分表
// 在数据库层面合并多个分表，统一排序和分页，性能更优
// 使用通用分表查询服务
func (s *OrderServiceImpl) queryMultipleTablesWithUnion(tableNames []string, filters OrderFilters, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	total, err := s.shardingQueryService.QueryMultipleTables(tableNames, filters, page, pageSize, &orders)
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

// queryMultipleTablesWithUnionForExport 使用 UNION ALL 查询多个分表（用于导出，不分页）
// 在数据库层面合并多个分表，统一排序，性能更优
// 使用通用分表查询服务
func (s *OrderServiceImpl) queryMultipleTablesWithUnionForExport(tableNames []string, filters OrderFilters) ([]models.Order, error) {
	var orders []models.Order
	err := s.shardingQueryService.QueryMultipleTablesForExport(tableNames, filters, &orders)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// queryMultipleTablesSequentially 逐个查询多个分表（用于商品名搜索等需要关联其他分表的场景）
func (s *OrderServiceImpl) queryMultipleTablesSequentially(tableNames []string, filters OrderFilters, page, pageSize int) ([]models.Order, int64, error) {
	var allOrders []models.Order
	var total int64

	// 应用排序
	orderBy := filters.OrderBy
	if orderBy == "" {
		orderBy = "created_at:desc"
	}

	// 逐个查询每个分表
	for _, tableName := range tableNames {
		// 检查表是否存在
		if !facades.Schema().HasTable(tableName) {
			continue
		}

		// 构建查询
		query := s.buildShardingQuery(tableName, filters)

		// 获取该分表的总数
		countQuery := s.buildShardingQuery(tableName, filters)
		tableTotal, err := countQuery.Count()
		if err != nil {
			continue
		}
		total += tableTotal

		// 查询数据（查询足够多的数据用于后续合并排序）
		query = s.applyOrderBy(query, orderBy)
		var orders []models.Order
		if err := query.Limit(page * pageSize).Find(&orders); err != nil {
			continue
		}
		allOrders = append(allOrders, orders...)
	}

	// 对合并后的数据进行排序
	s.sortOrders(allOrders, orderBy)

	// 应用分页
	offset := (page - 1) * pageSize
	if offset >= len(allOrders) {
		return []models.Order{}, total, nil
	}
	end := offset + pageSize
	if end > len(allOrders) {
		end = len(allOrders)
	}

	return allOrders[offset:end], total, nil
}

// sortOrders 对订单列表进行排序
func (s *OrderServiceImpl) sortOrders(orders []models.Order, orderBy string) {
	parts := strings.Split(orderBy, ":")
	field := "created_at"
	desc := true
	if len(parts) == 2 {
		field = parts[0]
		desc = strings.ToLower(parts[1]) == "desc"
	}

	// 使用 sort.Slice 进行排序
	sort.Slice(orders, func(i, j int) bool {
		var less bool
		switch field {
		case "id":
			less = orders[i].ID < orders[j].ID
		case "amount":
			less = orders[i].Amount < orders[j].Amount
		case "created_at":
			less = orders[i].CreatedAt.ToDateTimeString() < orders[j].CreatedAt.ToDateTimeString()
		case "updated_at":
			less = orders[i].UpdatedAt.ToDateTimeString() < orders[j].UpdatedAt.ToDateTimeString()
		default:
			less = orders[i].CreatedAt.ToDateTimeString() < orders[j].CreatedAt.ToDateTimeString()
		}
		if desc {
			return !less
		}
		return less
	})
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

	// 使用优化的 count 查询（使用公用的阈值配置）
	countOptimizer := utils.NewCountOptimizer(OrderCountThreshold, "order")
	whereClause, whereArgs := s.buildOrderWhereClause(filters)
	total, _, err := countOptimizer.OptimizedCountWithTable(tableName, whereClause, whereArgs...)
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

	if len(orders) == 0 {
		return []OrderWithDetails{}, total, nil
	}

	result := make([]OrderWithDetails, len(orders))

	// 按分表分组订单ID
	orderIDsByTable := make(map[string][]uint)
	orderIndexByID := make(map[uint]int)

	for i, order := range orders {
		// 将订单转换为 OrderWithDetails
		result[i] = OrderWithDetails{
			Order:   order,
			Details: []models.OrderDetail{},
		}

		// 根据订单的 created_at 确定详情分表
		timeStr := order.CreatedAt.ToDateTimeString()
		createdAt, _ := utils.ParseDateTimeUTC(timeStr)

		// 获取详情分表名
		detailTableName := utils.GetShardingTableName("order_details", createdAt)

		// 按分表分组订单ID
		if orderIDsByTable[detailTableName] == nil {
			orderIDsByTable[detailTableName] = []uint{}
		}
		orderIDsByTable[detailTableName] = append(orderIDsByTable[detailTableName], order.ID)
		orderIndexByID[order.ID] = i
	}

	// 按分表批量查询订单详情
	for tableName, orderIDs := range orderIDsByTable {
		if len(orderIDs) == 0 {
			continue
		}
		orderIDsAny := make([]any, len(orderIDs))
		for i, id := range orderIDs {
			orderIDsAny[i] = id
		}

		var details []models.OrderDetail
		if err := facades.Orm().Query().Table(tableName).
			WhereIn("order_id", orderIDsAny).
			Find(&details); err == nil {
			// 将详情分配到对应的订单
			for _, detail := range details {
				if index, ok := orderIndexByID[detail.OrderID]; ok {
					result[index].Details = append(result[index].Details, detail)
				}
			}
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
// 使用 UNION ALL 优化，在数据库层面合并和排序
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

	// 如果只有一个分表，直接查询
	if len(tableNames) == 1 {
		query := s.buildShardingQuery(tableNames[0], filters)
		orderBy := filters.OrderBy
		if orderBy == "" {
			orderBy = "created_at:desc"
		}
		query = s.applyOrderBy(query, orderBy)

		var orders []models.Order
		if err := query.Find(&orders); err != nil {
			return nil, apperrors.ErrQueryFailed.WithError(err)
		}
		return orders, nil
	}

	// 多个分表：使用 UNION ALL 在数据库层面合并
	return s.queryMultipleTablesWithUnionForExport(tableNames, filters)
}

// GetAllOrdersWithDetailsForExport 获取所有订单及详情用于导出（限制不超过3个月，不分页）
// 使用 UNION ALL 优化，在数据库层面合并和排序
// 优化：使用批量查询避免 N+1 查询问题
func (s *OrderServiceImpl) GetAllOrdersWithDetailsForExport(filters OrderFilters) ([]OrderWithDetails, error) {
	// 先获取所有订单（使用优化的 UNION ALL 方法）
	allOrders, err := s.GetAllOrdersForExport(filters)
	if err != nil {
		return nil, err
	}

	if len(allOrders) == 0 {
		return []OrderWithDetails{}, nil
	}

	// 初始化结果
	result := make([]OrderWithDetails, len(allOrders))

	// 按分表分组订单ID，避免 N+1 查询
	orderIDsByTable := make(map[string][]uint)
	orderIndexByID := make(map[uint]int)

	for i, order := range allOrders {
		result[i] = OrderWithDetails{
			Order:   order,
			Details: []models.OrderDetail{},
		}

		// 根据订单的 created_at 确定详情分表
		timeStr := order.CreatedAt.ToDateTimeString()
		createdAt, _ := utils.ParseDateTimeUTC(timeStr)

		// 获取详情分表名
		detailTableName := utils.GetShardingTableName("order_details", createdAt)

		// 按分表分组订单ID
		if orderIDsByTable[detailTableName] == nil {
			orderIDsByTable[detailTableName] = []uint{}
		}
		orderIDsByTable[detailTableName] = append(orderIDsByTable[detailTableName], order.ID)
		orderIndexByID[order.ID] = i
	}

	// 按分表批量查询订单详情（避免 N+1 查询）
	for tableName, orderIDs := range orderIDsByTable {
		if len(orderIDs) == 0 {
			continue
		}

		// 将 []uint 转换为 []any
		orderIDsAny := make([]any, len(orderIDs))
		for i, id := range orderIDs {
			orderIDsAny[i] = id
		}

		var details []models.OrderDetail
		if err := facades.Orm().Query().Table(tableName).
			WhereIn("order_id", orderIDsAny).
			Find(&details); err == nil {
			// 将详情分配到对应的订单
			for _, detail := range details {
				if index, ok := orderIndexByID[detail.OrderID]; ok {
					result[index].Details = append(result[index].Details, detail)
				}
			}
		}
	}

	return result, nil
}

// UpdateOrderStatus 更新订单状态（已废弃，保留以兼容旧接口）
func (s *OrderServiceImpl) UpdateOrderStatus(orderID uint, orderTime time.Time, status string) error {
	return s.UpdateOrder(orderID, orderTime, status, "")
}

// UpdateOrder 更新订单（状态和备注）
// 如果提供了订单号，优先使用订单号查找订单（更高效，可直接定位分表）
func (s *OrderServiceImpl) UpdateOrder(orderID uint, orderTime time.Time, status string, remark string, orderNo ...string) error {
	// 先查找订单获取 created_at
	// 如果提供了订单号，优先使用订单号查找
	var order *models.Order
	var err error
	if len(orderNo) > 0 && orderNo[0] != "" {
		order, err = s.findOrderByID(orderID, orderNo[0])
	} else {
		order, err = s.findOrderByID(orderID)
	}
	if err != nil {
		return err
	}

	// 使用订单的 created_at 确定分表
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := utils.ParseDateTimeUTC(timeStr)
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
// 如果提供了订单号，优先使用订单号查找订单（更高效，可直接定位分表）
func (s *OrderServiceImpl) DeleteOrder(orderID uint, orderTime time.Time, orderNo ...string) error {
	// 先查找订单获取 created_at（只查询未删除的订单）
	// 如果提供了订单号，优先使用订单号查找
	var order *models.Order
	var err error
	if len(orderNo) > 0 && orderNo[0] != "" {
		order, err = s.findOrderByID(orderID, orderNo[0])
	} else {
		order, err = s.findOrderByID(orderID)
	}
	if err != nil {
		return err
	}

	// 使用订单的 created_at 确定分表（将 carbon.DateTime 转换为 time.Time）
	// 通过格式化字符串再解析的方式转换
	timeStr := order.CreatedAt.ToDateTimeString()
	createdAt, _ := utils.ParseDateTimeUTC(timeStr)

	// 软删除订单详情
	detailTableName := utils.GetShardingTableName("order_details", createdAt)
	if _, err := facades.Orm().Query().Table(detailTableName).Where("order_id", orderID).Delete(&models.OrderDetail{}); err != nil {
		errorlog.Record(context.Background(), "order", "删除订单详情失败", map[string]any{
			"order_id": orderID,
			"error":    err.Error(),
		}, "删除订单详情失败: %v", err)
		return apperrors.ErrDeleteOrderDetailFailed.WithError(err)
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
		return nil, apperrors.ErrOrderNotFound
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
		return nil, apperrors.ErrOrderNotFound
	}

	// 使用解析的年月确定分表
	tableName := utils.GetShardingTableName("orders", parsedTime)

	// 直接查询对应的分表
	var order models.Order
	if err := facades.Orm().Query().Table(tableName).Where("order_no", orderNo).First(&order); err == nil {
		return &order, nil
	}

	return nil, apperrors.ErrOrderNotFound
}

// GetOrdersCountInYear 获取最近一年的订单总数（用于仪表盘统计）
func (s *OrderServiceImpl) GetOrdersCountInYear() (int64, error) {
	// 计算最近一年的时间范围
	now := time.Now().UTC()
	startTime := now.AddDate(-1, 0, 0) // 一年前
	endTime := now

	// 获取需要查询的所有分表
	tableNames := utils.GetShardingTableNames("orders", startTime, endTime)
	if len(tableNames) == 0 {
		return 0, nil
	}

	// 构建 WHERE 条件（只包含时间范围）
	whereClause := "created_at >= ? AND created_at <= ?"
	whereArgs := []any{startTime, endTime}

	// 使用优化的 count 查询（阈值：100000）
	countOptimizer := utils.NewCountOptimizer(OrderCountThreshold, "order")
	var total int64

	// 分别对每个分表执行 COUNT，然后相加
	for _, tableName := range tableNames {
		// 检查表是否存在
		if !facades.Schema().HasTable(tableName) {
			continue
		}

		// 使用优化的 count 查询
		tableTotal, _, err := countOptimizer.OptimizedCountWithTable(tableName, whereClause, whereArgs...)
		if err != nil {
			errorlog.Record(context.Background(), "order", "查询分表总数失败", map[string]any{
				"table_name": tableName,
				"error":      err.Error(),
			}, "查询分表 %s 总数失败: %v", tableName, err)
			// 如果某个分表查询失败，继续查询其他分表，但记录错误
			continue
		}
		total += tableTotal
	}

	return total, nil
}
