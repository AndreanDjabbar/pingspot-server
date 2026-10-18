package model

type ReportVote struct {
    ID        uint         `json:"id" gorm:"primaryKey"`
    ReportID  uint         `json:"reportId" gorm:"not null;uniqueIndex:idx_user_report_vote"`
    UserID    uint         `json:"userId" gorm:"not null;uniqueIndex:idx_user_report_vote"`
    User      User         `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Report    Report       `json:"report" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    VoteType  ReportStatus `json:"voteType" gorm:"type:varchar(50);not null"`
    CreatedAt int64        `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt int64        `json:"updatedAt" gorm:"autoUpdateTime"`
}