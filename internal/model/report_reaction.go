package model

type ReactionType string

const (
	Like    ReactionType = "LIKE"
	Dislike ReactionType = "DISLIKE"
)

type ReportReaction struct {
	ID       uint         `gorm:"primaryKey"`
	UserID   uint         `gorm:"not null"`
	User	 User         `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Type	 ReactionType `gorm:"type:varchar(20);not null"`
	CreatedAt int64        `gorm:"autoCreateTime"`
	UpdatedAt int64        `gorm:"autoUpdateTime"`
	ReportID uint         `gorm:"not null"`
	Report   Report       `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}