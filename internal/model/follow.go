package model

type FollowingType string

const (
	FollowingTypeUser FollowingType = "user"
	FollowingTypeCommunity FollowingType = "community"
)

type Follow struct {
    ID             uint          `json:"id" gorm:"primaryKey"`
    FollowingID    uint          `json:"followingId" gorm:"not null"`
    FollowingType  FollowingType `json:"followingType" gorm:"not null"`
    FollowerUserID uint          `json:"followerUserId" gorm:"not null"`
    FollowerUser   User          `json:"followerUser" gorm:"foreignKey:FollowerUserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    CreatedAt      int64         `json:"createdAt" gorm:"autoCreateTime"`
}
