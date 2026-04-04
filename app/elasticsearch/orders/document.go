package esorders

import "goravel/app/models"

// OrderDocument 构建写入 ES 的订单文档（字段变更集中在此，便于与 mapping 对齐）。
func OrderDocument(order *models.Order, details []models.OrderDetail) map[string]any {
	names := make([]string, 0, len(details))
	for _, d := range details {
		names = append(names, d.ProductName)
	}
	return map[string]any{
		"id":            order.ID,
		"order_no":      order.OrderNo,
		"user_id":       order.UserID,
		"amount":        order.Amount,
		"status":        order.Status,
		"remark":        order.Remark,
		"created_at":    order.CreatedAt.ToDateTimeString(),
		"updated_at":    order.UpdatedAt.ToDateTimeString(),
		"product_names": names,
	}
}
