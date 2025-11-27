package feature

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/assert"

	"goravel/app/models"
	"goravel/app/services"
	"goravel/app/utils/traceid"
)

func TestSystemLogTraceIDPersistence(t *testing.T) {
	t.Parallel()

	uniqueTraceID := fmt.Sprintf("trace-%d", time.Now().UnixNano())
	log := models.SystemLog{
		Level:   "info",
		Module:  "trace_test",
		TraceID: uniqueTraceID,
		Message: "trace id persistence test",
		Context: `{"test":true}`,
	}

	err := facades.Orm().Query().Create(&log)
	assert.NoError(t, err)
	assert.NotZero(t, log.ID)

	t.Cleanup(func() {
		_, _ = facades.Orm().Query().Where("id", log.ID).Delete(&models.SystemLog{})
	})

	var stored models.SystemLog
	err = facades.Orm().Query().Where("id", log.ID).First(&stored)
	assert.NoError(t, err)
	assert.Equal(t, uniqueTraceID, stored.TraceID)

	var filtered models.SystemLog
	err = facades.Orm().Query().Where("trace_id", uniqueTraceID).First(&filtered)
	assert.NoError(t, err)
	assert.Equal(t, log.ID, filtered.ID)
}

func TestSystemLogServiceRecord(t *testing.T) {
	t.Parallel()

	service := services.NewSystemLogService()
	ctx, traceID := traceid.EnsureContext(context.Background())

	err := service.Record(ctx, "error", "unit-test", "trace helper test", map[string]any{"ok": true})
	assert.NoError(t, err)

	var stored models.SystemLog
	err = facades.Orm().Query().Where("trace_id", traceID).Order("id desc").First(&stored)
	assert.NoError(t, err)
	assert.Equal(t, "unit-test", stored.Module)

	t.Cleanup(func() {
		_, _ = facades.Orm().Query().Where("trace_id", traceID).Delete(&models.SystemLog{})
	})
}

