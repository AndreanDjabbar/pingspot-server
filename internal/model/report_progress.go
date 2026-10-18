package model

type ReportProgress struct {
    ID          uint         `json:"id" gorm:"primaryKey"`
    ReportID    uint         `json:"reportId" gorm:"not null"`
    UserID      uint         `json:"userId" gorm:"not null"`
    User        User         `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Report      Report       `json:"report" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Status      ReportStatus `json:"status" gorm:"type:varchar(50);not null"`
    Notes       string       `json:"notes" gorm:"type:text"`
    Attachment1 *string      `json:"attachment1,omitempty" gorm:"size:255"`
    Attachment2 *string      `json:"attachment2,omitempty" gorm:"size:255"`
    CreatedAt   int64        `json:"createdAt" gorm:"autoCreateTime"`
}