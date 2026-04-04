package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/goravel/framework/contracts/config"
)

// NewElasticsearchClient 使用框架配置创建 ES 客户端（按连接名）；创建后会 Ping 校验。
func NewElasticsearchClient(cfg config.Config, connectionName string) (*elasticsearch.Client, error) {
	if connectionName == "" {
		connectionName = cfg.GetString("elasticsearch.default", "default")
	}

	base := fmt.Sprintf("elasticsearch.connections.%s", connectionName)
	urlsStr := cfg.GetString(base+".urls", "")
	cloudID := cfg.GetString(base+".cloud_id", "")
	if cloudID == "" && strings.TrimSpace(urlsStr) == "" {
		return nil, fmt.Errorf("elasticsearch [%s]: urls 或 cloud_id 至少配置一项", connectionName)
	}

	esCfg := elasticsearch.Config{
		Addresses: splitElasticsearchURLs(urlsStr),
		CloudID:   cloudID,
		Username:  cfg.GetString(base+".username", ""),
		Password:  cfg.GetString(base+".password", ""),
		APIKey:    cfg.GetString(base+".api_key", ""),
	}

	if cfg.GetBool(base+".insecure_skip_verify", false) {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
		esCfg.Transport = t
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch [%s] 创建客户端失败: %w", connectionName, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("elasticsearch [%s] Ping 失败: %w", connectionName, err)
	}
	if res != nil {
		_ = res.Body.Close()
	}
	if res != nil && res.IsError() {
		return nil, fmt.Errorf("elasticsearch [%s] Ping 返回错误: %s", connectionName, res.Status())
	}
	return client, nil
}

func splitElasticsearchURLs(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		u := strings.TrimSpace(p)
		if u != "" {
			out = append(out, u)
		}
	}
	return out
}

// ElasticsearchIndexName 返回带 elasticsearch.index_prefix 的完整索引名。
func ElasticsearchIndexName(cfg config.Config, name string) string {
	return cfg.GetString("elasticsearch.index_prefix", "") + name
}
