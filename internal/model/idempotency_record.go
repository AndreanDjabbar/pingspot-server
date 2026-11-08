package model

import "time"

type IdempotencyRecord struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;uniqueIndex:idx_idempotency_scope" json:"userId"`
	IdempotencyKey string     `gorm:"size:255;not null;uniqueIndex:idx_idempotency_scope" json:"idempotencyKey"`
	Method         string     `gorm:"size:10;not null;uniqueIndex:idx_idempotency_scope" json:"method"`
	Route          string     `gorm:"size:255;not null;uniqueIndex:idx_idempotency_scope" json:"route"`
	RequestHash    string     `gorm:"size:64;not null" json:"requestHash"`
	Status         string     `gorm:"size:20;not null" json:"status"`
	ResponseStatus *int       `gorm:"default:null" json:"responseStatus"`
	ResponseBody   *[]byte    `gorm:"type:jsonb;default:null" json:"responseBody"`
	CreatedAt      time.Time  `json:"createdAt"`
	CompletedAt    *time.Time `json:"completedAt"`
}
