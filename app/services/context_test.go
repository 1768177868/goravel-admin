package services_test

import (
	"context"
	"testing"

	"goravel/app/services"
)

func TestNewAdminServiceImpl_AcceptsContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	svc := services.NewAdminServiceImpl(ctx)
	if svc == nil {
		t.Fatal("expected admin service instance")
	}
}

func TestNewOrderService_AcceptsContext(t *testing.T) {
	svc := services.NewOrderService(context.Background())
	if svc == nil {
		t.Fatal("expected order service instance")
	}
}
