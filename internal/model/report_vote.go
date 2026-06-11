package model

type ReportVote struct {
	ID       uint   `gorm:"primaryKey"`
	ReportID  uint         `gorm:"not null;uniqueIndex:idx_user_report_vote"`
    UserID    uint         `gorm:"not null;uniqueIndex:idx_user_report_vote"`
	User     User   `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Report   Report `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	VoteType ReportStatus `gorm:"type:varchar(50);not null"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}