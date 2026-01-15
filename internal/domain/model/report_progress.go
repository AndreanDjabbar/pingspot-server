package model

type ReportProgress struct {
	ID        uint   `gorm:"primaryKey"`
	ReportID  uint   `gorm:"not null"`
	UserID   uint         `gorm:"not null"`
	User	 User         `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Report    Report `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Status ReportStatus `gorm:"type:varchar(50);not null"`
	Notes     string `gorm:"type:text"`
	Attachment1 *string `gorm:"size:255"`
	Attachment2 *string `gorm:"size:255"`
	CreatedAt int64  `gorm:"autoCreateTime"`
}