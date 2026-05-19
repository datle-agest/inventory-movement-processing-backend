package gormc

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// GormTxManager cấu trúc triển khai TxManager dùng chung
type GormTxManager struct {
	db *gorm.DB
}

func NewGormTxManager(db *gorm.DB) *GormTxManager {
	return &GormTxManager{db: db}
}

// WithTx bọc nghiệp vụ trong Transaction của GORM
func (m *GormTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Đút instance Transaction vào context mới
		txCtx := context.WithValue(ctx, txKey{}, tx)

		// Thực thi hàm callback nghiệp vụ
		return fn(txCtx)
	})
}

// GetDB lấy đúng DB hoặc Tx từ context ra để dùng
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}
