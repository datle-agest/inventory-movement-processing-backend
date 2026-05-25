package service

import (
    "context"
    "inventory-movement-processing/internal/item/entity"
    "inventory-movement-processing/pkg/core"
    "inventory-movement-processing/pkg/logger"
)

// --- itemRepository mock ---
type mockItemRepo struct {
    getItemFn          func(ctx context.Context, id int32) (*entity.Item, error)
    listItemFn         func(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error)
    createItemFn       func(ctx context.Context, item entity.Item) (*entity.Item, error)
    listLowStockFn     func(ctx context.Context) ([]*entity.Item, error)
    updateStockFn      func(ctx context.Context, itemID int32, newStock int32) error
    getItemForUpdateFn func(ctx context.Context, id int32) (*entity.Item, error)
}

func (m *mockItemRepo) GetItem(ctx context.Context, id int32) (*entity.Item, error) {
    return m.getItemFn(ctx, id)
}

func (m *mockItemRepo) ListItem(ctx context.Context, filter *entity.ItemFilter, paging *core.Pagination) ([]entity.Item, error) {
    return m.listItemFn(ctx, filter, paging)
}

func (m *mockItemRepo) CreateItem(ctx context.Context, item entity.Item) (*entity.Item, error) {
    return m.createItemFn(ctx, item)
}

func (m *mockItemRepo) ListLowStockItems(ctx context.Context) ([]*entity.Item, error) {
    return m.listLowStockFn(ctx)
}

func (m *mockItemRepo) UpdateStock(ctx context.Context, itemID int32, newStock int32) error {
    return m.updateStockFn(ctx, itemID, newStock)
}

func (m *mockItemRepo) GetItemForUpdate(ctx context.Context, id int32) (*entity.Item, error) {
    return m.getItemForUpdateFn(ctx, id)
}

// --- Logger mock ---
type mockLogger struct{}

func (m *mockLogger) Debug(args ...interface{}) {}
func (m *mockLogger) Info(args ...interface{})  {}
func (m *mockLogger) Warn(args ...interface{})  {}
func (m *mockLogger) Error(args ...interface{}) {}

func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Infof(format string, args ...interface{})  {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}

func (m *mockLogger) With(key string, value interface{}) logger.Logger {
    return m
}

func (m *mockLogger) WithFields(fields logger.Fields) logger.Logger {
    return m
}

// Helpers

func newService(
    repo itemRepository,
    log logger.Logger,
) *itemService {
    return NewItemService(repo, log)
}