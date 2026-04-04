package esorders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/facades"

	"goravel/app/binding"
	"goravel/app/clients"
)

// MyOrderSearchHit 单条命中（与 OrderDocument 字段对齐，供 C 端展示）。
type MyOrderSearchHit struct {
	ID           uint     `json:"id"`
	OrderNo      string   `json:"order_no"`
	Amount       float64  `json:"amount"`
	Status       int      `json:"status"`
	Remark       string   `json:"remark"`
	CreatedAt    string   `json:"created_at"`
	ProductNames []string `json:"product_names"`
}

type esSearchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// SearchMyOrders 在 ES 中限定 user_id，按关键词多字段检索；keyword 为空时仅列出该用户订单（按 created_at 降序）。
func SearchMyOrders(ctx context.Context, userID uint, keyword string, page, pageSize int) (total int64, items []MyOrderSearchHit, err error) {
	raw, err := facades.App().Make(binding.ElasticsearchClient)
	if err != nil {
		return 0, nil, err
	}
	es, ok := raw.(*elasticsearch.Client)
	if !ok || es == nil {
		return 0, nil, fmt.Errorf("elasticsearch client not available")
	}

	short := facades.Config().GetString("elasticsearch.orders_index", "orders")
	index := clients.ElasticsearchIndexName(facades.Config(), short)

	boolQ := map[string]any{
		"filter": []any{
			map[string]any{"term": map[string]any{"user_id": userID}},
		},
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		boolQ["must"] = []any{
			map[string]any{
				"multi_match": map[string]any{
					"query":  kw,
					"type":   "best_fields",
					"fields": []string{"order_no^2", "product_names", "remark"},
				},
			},
		}
	}

	body := map[string]any{
		"query": map[string]any{"bool": boolQ},
		"from":  (page - 1) * pageSize,
		"size":  pageSize,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}

	res, err := es.Search(
		es.Search.WithContext(ctx),
		es.Search.WithIndex(index),
		es.Search.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, nil, err
	}
	if res.IsError() {
		return 0, nil, fmt.Errorf("elasticsearch search: %s: %s", res.Status(), string(b))
	}

	var parsed esSearchResponse
	if err := json.Unmarshal(b, &parsed); err != nil {
		return 0, nil, err
	}

	total = parsed.Hits.Total.Value
	items = make([]MyOrderSearchHit, 0, len(parsed.Hits.Hits))
	for _, h := range parsed.Hits.Hits {
		var hit MyOrderSearchHit
		if err := json.Unmarshal(h.Source, &hit); err != nil {
			continue
		}
		items = append(items, hit)
	}
	return total, items, nil
}
