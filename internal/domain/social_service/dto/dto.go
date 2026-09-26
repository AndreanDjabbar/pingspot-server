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
	FollowID uint   `json:"followID"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	ProfilePicture *string `json:"profilePicture"`
	Status   string `json:"status"`
	Relation string `json:"relation"`
}

type GetConnectionsRepositoryResponse struct {
	User *model.User `json:"user"`
	Follow *model.Follow `json:"follow"`
}