package model

type ReportSaved struct {
	ID       uint         `gorm:"primaryKey"`
	UserID   uint         `gorm:"not null"`
	User	 User         `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt int64        `gorm:"autoCreateTime"`
	ReportID uint         `gorm:"not null"`
	Report   Report       `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}