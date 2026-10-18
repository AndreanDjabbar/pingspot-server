package model

type ReactionType string

const (
	Like    ReactionType = "LIKE"
	Dislike ReactionType = "DISLIKE"
)

type ReportReaction struct {
    ID        uint         `json:"id" gorm:"primaryKey"`
    UserID    uint         `json:"userId" gorm:"not null"`
    User      User         `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Type      ReactionType `json:"type" gorm:"type:varchar(20);not null"`
    CreatedAt int64        `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt int64        `json:"updatedAt" gorm:"autoUpdateTime"`
    ReportID  uint         `json:"reportId" gorm:"not null"`
    Report    Report       `json:"report" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}