package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
)

// ElasticsearchIndexJSON 写入/覆盖指定 _id 的文档（refresh=true）。
func ElasticsearchIndexJSON(ctx context.Context, es *elasticsearch.Client, index, documentID string, body []byte) error {
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
		return fmt.Errorf("es index: %s: %s", res.Status(), string(b))
	}
	return nil
}

// ElasticsearchIndexValue 将任意可 JSON 序列化的值写入 ES。
func ElasticsearchIndexValue(ctx context.Context, es *elasticsearch.Client, index, documentID string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return ElasticsearchIndexJSON(ctx, es, index, documentID, body)
}

// ElasticsearchDeleteDocument 按文档 ID 删除；404 视为成功。
func ElasticsearchDeleteDocument(ctx context.Context, es *elasticsearch.Client, index, documentID string) error {
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
		return fmt.Errorf("es delete: %s: %s", res.Status(), string(b))
	}
	return nil
}
