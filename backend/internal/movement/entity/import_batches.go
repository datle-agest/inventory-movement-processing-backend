package entity

import (
	"database/sql/driver"
	"encoding/json"
	"inventory-movement-processing/pkg/core"
	"strconv"
	"strings"
	"time"
)

type BatchStatus string

const (
	BatchStatusProcessing BatchStatus = "PROCESSING"
	BatchStatusCompleted  BatchStatus = "COMPLETED"
	BatchStatusFailed     BatchStatus = "FAILED"
)

type ErrorLog map[string]int32 // {"invalid_sku": 500, "duplicate": 1000, ...}

func (el ErrorLog) Value() (driver.Value, error) {
	return json.Marshal(el)
}

func (el *ErrorLog) Scan(value interface{}) error {
	bytes, _ := value.([]byte)
	return json.Unmarshal(bytes, &el)
}

type ImportBatch struct {
	core.SQLModel
	FileName      string      `json:"file_name"      gorm:"column:file_name;type:varchar(255);not null"`
	Status        BatchStatus `json:"status"         gorm:"column:status;type:varchar(20);default:'PROCESSING'"`
	TotalRows     int32       `json:"total_rows"     gorm:"column:total_rows;not null"`
	ProcessedRows int32       `json:"processed_rows" gorm:"column:processed_rows;not null;default:0"`
	SuccessRows   int32       `json:"success_rows"   gorm:"column:success_rows;not null;default:0"`
	FailedRows    int32       `json:"failed_rows"    gorm:"column:failed_rows;not null;default:0"`
	ErrorLog      ErrorLog    `json:"error_log"      gorm:"column:error_log;type:jsonb;default:'{}'"`
	StartedAt     *time.Time  `json:"started_at"     gorm:"column:started_at"`
	CompletedAt   *time.Time  `json:"completed_at"   gorm:"column:completed_at"`
}

func (ImportBatch) TableName() string {
	return "import_batches"
}

func (ib *ImportBatch) Validate() error {
	if ib.FileName == "" {
		return ErrFileNameEmpty
	}
	if ib.TotalRows <= 0 {
		return ErrInvalidTotalRows
	}
	return nil
}

func (ib *ImportBatch) AddError(errorType string, count int32) {
	if ib.ErrorLog == nil {
		ib.ErrorLog = make(ErrorLog)
	}
	ib.ErrorLog[errorType] += count
}

func (ib *ImportBatch) GetErrorSummary() string {
	summary := ""
	for errorType, count := range ib.ErrorLog {
		summary += errorType + ": " + strconv.Itoa(int(count)) + ", "
	}
	return strings.TrimSuffix(summary, ", ")
}
