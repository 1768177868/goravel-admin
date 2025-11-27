package services

import (
	"context"
	"encoding/json"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"

	"goravel/app/models"
	"goravel/app/utils/traceid"
)

type SystemLogService interface {
	RecordHTTP(ctx http.Context, level, module, message string, attributes map[string]any) error
	Record(ctx context.Context, level, module, message string, attributes map[string]any) error
}

type SystemLogServiceImpl struct{}

func NewSystemLogService() *SystemLogServiceImpl {
	return &SystemLogServiceImpl{}
}

func (s *SystemLogServiceImpl) RecordHTTP(ctx http.Context, level, module, message string, attributes map[string]any) error {
	var contextJSON string
	if len(attributes) > 0 {
		if data, err := json.Marshal(attributes); err == nil {
			contextJSON = string(data)
		}
	}

	traceID := traceid.FromHTTPContext(ctx)
	if traceID == "" {
		traceID = traceid.EnsureHTTPContext(ctx, "")
	}

	log := models.SystemLog{
		Level:     level,
		Module:    module,
		TraceID:   traceID,
		Message:   message,
		Context:   contextJSON,
		IP:        ctx.Request().Ip(),
		UserAgent: ctx.Request().Header("User-Agent", ""),
	}

	return facades.Orm().Query().Create(&log)
}

func (s *SystemLogServiceImpl) Record(ctx context.Context, level, module, message string, attributes map[string]any) error {
	if ctx == nil {
		ctx = context.Background()
	}

	var contextJSON string
	if len(attributes) > 0 {
		if data, err := json.Marshal(attributes); err == nil {
			contextJSON = string(data)
		}
	}

	var traceID string
	if id, ok := ctx.Value(traceid.ContextKey).(string); ok {
		traceID = id
	}
	if traceID == "" {
		var newCtx context.Context
		newCtx, traceID = traceid.EnsureContext(ctx)
		ctx = newCtx
	}

	log := models.SystemLog{
		Level:   level,
		Module:  module,
		TraceID: traceID,
		Message: message,
		Context: contextJSON,
	}

	return facades.Orm().Query().Create(&log)
}
