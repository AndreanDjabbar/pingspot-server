package dto

import (
	"pingspot/internal/domain/model"
	userDTO "pingspot/internal/domain/userService/dto"
)

type CreateReportResponse struct {
	Report         model.Report         `json:"report"`
	ReportLocation model.ReportLocation `json:"reportLocation"`
	ReportImages   model.ReportImage    `json:"reportImages"`
}

type EditReportResponse struct {
	Report         model.Report         `json:"report"`
	ReportLocation model.ReportLocation `json:"reportLocation"`
	ReportImages   model.ReportImage    `json:"reportImages"`
}

type GetReportsResponse struct {
	Reports 				[]Report                `json:"reports"`
	TotalCounts             *TotalReportCount           `json:"totalCounts,omitempty"`
}

type GetReportResponse struct {
	Report      Report       `json:"report"`
}

type ReactReportResponse struct {
	ReportID     uint   `json:"reportID"`
	UserID       uint   `json:"userID"`
	ReactionType string `json:"reactionType"`
	CreatedAt    int64  `json:"createdAt"`
	UpdatedAt    int64  `json:"updatedAt"`
}

type UploadProgressReportResponse struct {
	ReportID    uint    `json:"reportID"`
	Status      string  `json:"status"`
	Notes       *string `json:"notes"`
	Attachment1 *string `json:"attachment1"`
	Attachment2 *string `json:"attachment2"`
	CreatedAt   int64   `json:"createdAt"`
	LastUpdatedProgressAt   *int64   `json:"lastUpdatedProgressAt,omitempty"`
}

type GetVoteReportResponse struct {
	ID        uint               `json:"id"`
	ReportID  uint               `json:"reportID"`
	ReportStatus model.ReportStatus `json:"reportStatus"`
	UserID    uint               `json:"userID"`
	VoteType  model.ReportStatus `json:"voteType"`
	CreatedAt int64              `json:"createdAt"`
	UpdatedAt int64              `json:"updatedAt"`
	LastUpdatedBy   *string            `json:"lastUpdatedBy,omitempty"`
	LastUpdatedProgressAt   *int64            `json:"lastUpdatedProgressAt,omitempty"`
}

type GetProgressReportResponse struct {
	ID          uint    `json:"id"`
	ReportID    uint    `json:"reportID"`
	Status      string  `json:"status"`
	Notes       *string `json:"notes"`
	Attachment1 *string `json:"attachment1"`
	Attachment2 *string `json:"attachment2"`
	CreatedAt   int64   `json:"createdAt"`
}

type CreateReportCommentResponse struct {
	CommentID      string `json:"commentID"`
	ReportID       uint   `json:"reportID"`
	UserID         uint   `json:"userID"`
	Content        *string `json:"content,omitempty"`
	CreatedAt      int64  `json:"createdAt"`
	ReplyTo 	  *userDTO.UserProfile `json:"replyTo,omitempty"`
	Media 	   	  *CommentMedia `json:"media,omitempty"`
	ThreadRootID   *string `json:"threadRootID,omitempty"`
	ParentCommentID *string `json:"parentCommentID,omitempty"`
}

type GetReportCommentsResponse struct {
	Comments []*Comment `json:"comments"`
	TotalCounts int64    `json:"totalCounts"`
	HasMore     bool      `json:"hasMore"`
}

type GetReportCommentRepliesResponse struct {
	Replies     []*CommentReply `json:"replies"`
	TotalCounts int64           `json:"totalCounts"`	
	HasMore     bool           `json:"hasMore"`
}

type GetReportStatisticsResponse struct {
	TotalReports 	  int64 `json:"totalReports"`
	ReportsByStatus   map[string]int64 `json:"reportsByStatus"`
	MonthlyReportCounts map[string]int64 `json:"monthlyReportCounts"`
}