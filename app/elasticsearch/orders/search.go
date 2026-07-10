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
	"goravel/app/dto"
)

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
// createdAtGTE、createdAtLTE 为可选的 created_at 字符串边界（与 OrderDocument 中格式一致），用于 range 过滤。
func SearchMyOrders(ctx context.Context, userID uint, keyword string, page, pageSize int, createdAtGTE, createdAtLTE *string) (total int64, items []dto.OrderSearchListItem, err error) {
	raw, err := facades.App().Make(binding.ElasticsearchClient)
	if err != nil {
		return 0, nil, err
	}
	es, ok := raw.(*elasticsearch.Client)
	if !ok || es == nil {
		return 0, nil, fmt.Errorf("elasticsearch client not available")
	}

	index := OrdersIndexName()

	filters := []any{
		map[string]any{"term": map[string]any{"user_id": userID}},
	}
	if createdAtGTE != nil || createdAtLTE != nil {
		rng := map[string]any{}
		if createdAtGTE != nil {
			rng["gte"] = *createdAtGTE
		}
		if createdAtLTE != nil {
			rng["lte"] = *createdAtLTE
		}
		filters = append(filters, map[string]any{
			"range": map[string]any{"created_at": rng},
		})
	}

	boolQ := map[string]any{
		"filter": filters,
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
	items = make([]dto.OrderSearchListItem, 0, len(parsed.Hits.Hits))
	for _, h := range parsed.Hits.Hits {
		var hit dto.OrderSearchListItem
		if err := json.Unmarshal(h.Source, &hit); err != nil {
			continue
		}
		items = append(items, hit)
	}
	return total, items, nil
}
