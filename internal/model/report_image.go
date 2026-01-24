package model

type ReportImage struct {
	ID        uint    `gorm:"primaryKey"`
	ReportID  uint    `gorm:"not null"`
	Report    Report  `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Image1URL *string `gorm:"size:255"`
	Image2URL *string `gorm:"size:255"`
	Image3URL *string `gorm:"size:255"`
	Image4URL *string `gorm:"size:255"`
	Image5URL *string `gorm:"size:255"`
}