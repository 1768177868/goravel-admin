package esorders

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/facades"
)

// ResolveOrdersTextAnalyzers 解析订单索引中文本字段的分词器。
// auto（默认）：集群有 IK 则用 ik_max_word / ik_smart，否则不设 analyzer（ES 默认 standard）。
// standard：不设 analyzer。其它值：按名称使用；不可用时回退 standard，避免未装插件时建索引失败。
func ResolveOrdersTextAnalyzers(ctx context.Context, es *elasticsearch.Client) (indexAnalyzer, searchAnalyzer string) {
	mode := strings.ToLower(strings.TrimSpace(facades.Config().GetString("elasticsearch.orders_analyzer", "auto")))
	searchOverride := strings.TrimSpace(facades.Config().GetString("elasticsearch.orders_search_analyzer", ""))

	switch mode {
	case "", "standard", "default":
		return "", ""
	case "auto":
		if es != nil && analyzerAvailable(ctx, es, "ik_max_word") {
			search := "ik_smart"
			if searchOverride != "" {
				search = searchOverride
			}
			facades.Log().Info("elasticsearch orders index: using IK analyzers (ik_max_word / " + search + ")")
			return "ik_max_word", search
		}
		facades.Log().Info("elasticsearch orders index: IK not available, using default standard analyzer")
		return "", ""
	default:
		index := mode
		search := searchOverride
		if search == "" {
			if index == "ik_max_word" {
				search = "ik_smart"
			} else {
				search = index
			}
		}
		if es != nil && !analyzerAvailable(ctx, es, index) {
			facades.Log().Warningf("elasticsearch orders analyzer %q unavailable, fallback to standard", index)
			return "", ""
		}
		facades.Log().Infof("elasticsearch orders index: using analyzers (%s / %s)", index, search)
		return index, search
	}
}

func analyzerAvailable(ctx context.Context, es *elasticsearch.Client, analyzer string) bool {
	if es == nil || strings.TrimSpace(analyzer) == "" {
		return false
	}
	payload, err := json.Marshal(map[string]any{
		"analyzer": analyzer,
		"text":     "中文分词检测",
	})
	if err != nil {
		return false
	}
	res, err := es.Indices.Analyze(
		es.Indices.Analyze.WithContext(ctx),
		es.Indices.Analyze.WithBody(bytes.NewReader(payload)),
	)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	if res.IsError() {
		return false
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return false
	}
	var parsed struct {
		Tokens []struct {
			Token string `json:"token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return false
	}
	return len(parsed.Tokens) > 0
}

func textFieldMapping(indexAnalyzer, searchAnalyzer string) map[string]any {
	m := map[string]any{"type": "text"}
	if indexAnalyzer != "" {
		m["analyzer"] = indexAnalyzer
	}
	if searchAnalyzer != "" {
		m["search_analyzer"] = searchAnalyzer
	}
	return m
}
