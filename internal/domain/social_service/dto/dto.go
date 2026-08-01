package dto

import "pingspot/internal/model"

type Follow struct {
	ID             uint                `gorm:"primaryKey"`
	FollowingID    uint                `gorm:"not null"`
	FollowingType  model.FollowingType `gorm:"not null"`
	FollowerUserID uint                `gorm:"not null"`
	CreatedAt      int64               `gorm:"autoCreateTime"`
}