package entity

type CreateItemRequest struct {
	Name              string `json:"name" binding:"required,min=2,max=100"`
	SKU               string `json:"sku" binding:"required,min=3,max=50"`
	CurrentStock      int32  `json:"current_stock" binding:"gte=0"`
	LowStockThreshold int32  `json:"low_stock_threshold" binding:"gte=0"`
}
