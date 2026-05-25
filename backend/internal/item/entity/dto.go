package entity

type CreateItemRequest struct {
	Name              string `json:"name" validate:"required,min=2,max=100"`
	SKU               string `json:"sku" validate:"required,min=3,max=50"`
	CurrentStock      int32  `json:"current_stock" validate:"gte=0"`
	LowStockThreshold int32  `json:"low_stock_threshold" validate:"gte=0"`
}
