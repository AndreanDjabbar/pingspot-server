package model

type ReportImage struct {
    ID        uint   `json:"id" gorm:"primaryKey"`
    ReportID  uint   `json:"reportId" gorm:"not null"`
    Report    Report `json:"report" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Image1URL *string `json:"image1Url,omitempty" gorm:"size:255"`
    Image2URL *string `json:"image2Url,omitempty" gorm:"size:255"`
    Image3URL *string `json:"image3Url,omitempty" gorm:"size:255"`
    Image4URL *string `json:"image4Url,omitempty" gorm:"size:255"`
    Image5URL *string `json:"image5Url,omitempty" gorm:"size:255"`
}