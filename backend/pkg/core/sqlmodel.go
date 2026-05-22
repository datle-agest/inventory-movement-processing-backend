package core

import "time"

type SQLModel struct {
	ID        int32      `json:"id" gorm:"column:id;" db:"id" example:"1"`
	CreatedAt *time.Time `json:"created_at,omitempty" gorm:"column:created_at;" db:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at;" db:"updated_at" example:"2024-01-15T10:30:00Z"`
}

func NewSQLModel() SQLModel {
	now := time.Now().UTC()

	return SQLModel{
		ID:        0,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
}
