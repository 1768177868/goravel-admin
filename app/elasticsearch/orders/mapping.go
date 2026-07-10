package esorders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/facades"

	"goravel/app/binding"
	"goravel/app/clients"
)

var ensureIndexOnce sync.Once
var ensureIndexErr error

// OrdersIndexName 订单 ES 索引全名（含 prefix）。
func OrdersIndexName() string {
	short := facades.Config().GetString("elasticsearch.orders_index", "orders")
	return clients.ElasticsearchIndexName(facades.Config(), short)
}

func ordersIndexBody() map[string]any {
	return map[string]any{
		"mappings": map[string]any{
			"properties": map[string]any{
				"id":       map[string]any{"type": "long"},
				"order_no": map[string]any{"type": "keyword"},
				"user_id":  map[string]any{"type": "long"},
				"amount":   map[string]any{"type": "double"},
				"status":   map[string]any{"type": "keyword"},
				"remark":   map[string]any{"type": "text"},
				"created_at": map[string]any{
					"type":   "date",
					"format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||strict_date_optional_time||epoch_millis",
				},
				"updated_at": map[string]any{
					"type":   "date",
					"format": "yyyy-MM-dd HH:mm:ss||yyyy-MM-dd||strict_date_optional_time||epoch_millis",
				},
				"product_names": map[string]any{"type": "text"},
			},
		},
	}
}

// EnsureOrdersIndex 创建订单索引（若不存在）；进程内只执行一次，供写入路径调用。
func EnsureOrdersIndex(ctx context.Context) error {
	ensureIndexOnce.Do(func() {
		ensureIndexErr = initOrdersIndex(ctx)
	})
	return ensureIndexErr
}

// InitOrdersIndex 创建订单索引（若不存在）；供 artisan 命令显式初始化。
func InitOrdersIndex(ctx context.Context) error {
	return initOrdersIndex(ctx)
}

func initOrdersIndex(ctx context.Context) error {
	raw, err := facades.App().Make(binding.ElasticsearchClient)
	if err != nil {
		return err
	}
	es, ok := raw.(*elasticsearch.Client)
	if !ok || es == nil {
		return fmt.Errorf("elasticsearch client not available")
	}

	index := OrdersIndexName()
	res, err := es.Indices.Exists(
		[]string{index},
		es.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	if res != nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			return nil
		}
	}

	body, err := json.Marshal(ordersIndexBody())
	if err != nil {
		return err
	}
	createRes, err := es.Indices.Create(
		index,
		es.Indices.Create.WithContext(ctx),
		es.Indices.Create.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return err
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		b, _ := io.ReadAll(createRes.Body)
		// 并发创建时另一进程可能已创建成功
		if createRes.StatusCode == 400 && bytes.Contains(b, []byte("resource_already_exists")) {
			return nil
		}
		return fmt.Errorf("es create index %s: %s: %s", index, createRes.Status(), string(b))
	}
	return nil
}
