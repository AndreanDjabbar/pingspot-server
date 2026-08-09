package dto

import "pingspot/internal/model"

type Follow struct {
	ID             uint                `gorm:"primaryKey"`
	FollowingID    uint                `gorm:"not null"`
	FollowingType  model.FollowingType `gorm:"not null"`
	FollowerUserID uint                `gorm:"not null"`
	CreatedAt      int64               `gorm:"autoCreateTime"`
}

type UserConnection struct {
	UserID   uint   `json:"userID"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	ProfilePicture *string `json:"profilePicture"`
	Status   string `json:"status"`
	Relation string `json:"relation"`
}