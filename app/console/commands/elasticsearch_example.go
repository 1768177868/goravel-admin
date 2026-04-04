package commands

import (
	"context"
	"fmt"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"

	"goravel/app/binding"
	"goravel/app/services"
)

// ElasticsearchExample 演示 Ping / 写入 / 搜索 / 删除（需 ELASTICSEARCH_ENABLED=true）。
type ElasticsearchExample struct{}

func (r *ElasticsearchExample) Signature() string {
	return "es:example"
}

func (r *ElasticsearchExample) Description() string {
	return "Elasticsearch 示例：ping | index | search | delete（需 .env 开启 ELASTICSEARCH_ENABLED）"
}

func (r *ElasticsearchExample) Extend() command.Extend {
	return command.Extend{
		Category: "elasticsearch",
		Flags: []command.Flag{
			&command.StringFlag{
				Name:    "op",
				Aliases: []string{"o"},
				Usage:   "操作: ping（默认）| index | search | delete",
				Value:   "ping",
			},
			&command.StringFlag{
				Name:  "id",
				Usage: "index/delete 时的文档 ID，默认 demo-1",
				Value: "demo-1",
			},
			&command.StringFlag{
				Name:  "q",
				Usage: "search 时的关键词，默认 示例",
				Value: "示例",
			},
		},
	}
}

func (r *ElasticsearchExample) Handle(ctx console.Context) error {
	if !facades.Config().GetBool("elasticsearch.enabled", false) {
		ctx.Warning("未启用 Elasticsearch。请在 .env 设置 ELASTICSEARCH_ENABLED=true 并配置 ELASTICSEARCH_URLS 等变量。")
		return nil
	}

	if _, err := facades.App().Make(binding.ElasticsearchClient); err != nil {
		ctx.Error(fmt.Sprintf("无法解析 ES 客户端: %v", err))
		return err
	}

	svc := services.NewElasticsearchService()
	op := ctx.Option("op")
	c := context.Background()

	switch op {
	case "ping", "":
		if err := svc.Ping(c); err != nil {
			ctx.Error(err.Error())
			return err
		}
		ctx.Success("Ping 成功，集群可达")
		return nil

	case "index":
		index := svc.DemoIndex()
		doc := map[string]any{
			"title":       "Goravel ES 示例文档",
			"description": "这是一条用于演示索引与搜索的示例数据",
			"tags":        []string{"demo", "goravel"},
		}
		id := ctx.Option("id")
		if err := svc.IndexDocument(c, index, id, doc); err != nil {
			ctx.Error(err.Error())
			return err
		}
		ctx.Success(fmt.Sprintf("已写入索引 %s 文档 id=%s", index, id))
		ctx.Info("可执行: go run . artisan es:example --op=search")
		return nil

	case "search":
		index := svc.DemoIndex()
		raw, err := svc.SearchMatch(c, index, "title", ctx.Option("q"), 5)
		if err != nil {
			ctx.Error(err.Error())
			return err
		}
		ctx.Info(string(raw))
		return nil

	case "delete":
		index := svc.DemoIndex()
		id := ctx.Option("id")
		if err := svc.DeleteDocument(c, index, id); err != nil {
			ctx.Error(err.Error())
			return err
		}
		ctx.Success(fmt.Sprintf("已删除索引 %s 文档 id=%s（若不存在则忽略 404）", index, id))
		return nil

	default:
		ctx.Error("未知 --op，支持: ping | index | search | delete")
		return fmt.Errorf("invalid op")
	}
}
