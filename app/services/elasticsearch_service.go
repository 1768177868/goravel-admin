package services

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

// ElasticsearchService ES 读写与搜索封装；需在 ELASTICSEARCH_ENABLED=true 且容器已绑定客户端后使用。
// 使用示例：./artisan es:example ，或登录后 GET api/admin/elasticsearch/ping。
type ElasticsearchService struct{}

func NewElasticsearchService() *ElasticsearchService {
	return &ElasticsearchService{}
}

func (s *ElasticsearchService) client() (*elasticsearch.Client, error) {
	raw, err := facades.App().Make(binding.ElasticsearchClient)
	if err != nil {
		return nil, err
	}
	client, ok := raw.(*elasticsearch.Client)
	if !ok || client == nil {
		return nil, fmt.Errorf("elasticsearch 客户端未注册：请在 .env 设置 ELASTICSEARCH_ENABLED=true 并正确配置连接")
	}
	return client, nil
}

// DemoIndex 返回示例用的完整索引名（含 index_prefix）。
func (s *ElasticsearchService) DemoIndex() string {
	short := facades.Config().GetString("elasticsearch.demo_index", "goravel_demo")
	return clients.ElasticsearchIndexName(facades.Config(), short)
}

// Ping 集群连通性检查。
func (s *ElasticsearchService) Ping(ctx context.Context) error {
	es, err := s.client()
	if err != nil {
		return err
	}
	res, err := es.Ping(es.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	if res != nil {
		defer res.Body.Close()
	}
	if res != nil && res.IsError() {
		return fmt.Errorf("ping: %s", res.Status())
	}
	return nil
}

// IndexDocument 索引单条 JSON 文档（示例：写入任意 map）。
func (s *ElasticsearchService) IndexDocument(ctx context.Context, index, documentID string, doc any) error {
	es, err := s.client()
	if err != nil {
		return err
	}
	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	res, err := es.Index(
		index,
		bytes.NewReader(body),
		es.Index.WithContext(ctx),
		es.Index.WithDocumentID(documentID),
		es.Index.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index: %s: %s", res.Status(), string(b))
	}
	return nil
}

// SearchMatch 简单 match 查询，返回 ES 原始 JSON（便于前端或后续解析）。
func (s *ElasticsearchService) SearchMatch(ctx context.Context, index, field, query string, size int) (json.RawMessage, error) {
	if size <= 0 {
		size = 10
	}
	es, err := s.client()
	if err != nil {
		return nil, err
	}
	q := map[string]any{
		"query": map[string]any{
			"match": map[string]any{
				field: query,
			},
		},
		"size": size,
	}
	body, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}
	res, err := es.Search(
		es.Search.WithContext(ctx),
		es.Search.WithIndex(index),
		es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("search: %s: %s", res.Status(), strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

// DeleteDocument 按文档 ID 删除。
func (s *ElasticsearchService) DeleteDocument(ctx context.Context, index, documentID string) error {
	es, err := s.client()
	if err != nil {
		return err
	}
	res, err := es.Delete(
		index,
		documentID,
		es.Delete.WithContext(ctx),
		es.Delete.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != 404 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete: %s: %s", res.Status(), string(b))
	}
	return nil
}
