package postgres

import "time"

func (TaskRecord) TableName() string {
	return "tasks"
}

type TaskRecord struct {
	ID          int       `gorm:"primaryKey;autoIncrement"`
	Title       string    `gorm:"type:text;not null"`
	Description string    `gorm:"type:text;not null;default:''"`
	Done        bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}
