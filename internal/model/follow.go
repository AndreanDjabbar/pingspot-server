package model

type FollowingType string

const (
	FollowingTypeUser FollowingType = "user"
	FollowingTypeCommunity FollowingType = "community"
)

type Follow struct {
	ID             uint   `gorm:"primaryKey"`
	FollowingID uint   `gorm:"not null"`
	FollowingType FollowingType `gorm:"not null"`
	FollowerUserID  uint   `gorm:"not null"`
	FollowerUser  User   `gorm:"foreignKey:FollowerUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
}
