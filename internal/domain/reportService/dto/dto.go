package dto

import userDTO "pingspot/internal/domain/userService/dto"

type TotalReportCount struct {
	TotalReports               int64 `json:"totalReports"`
	TotalInfrastructureReports int64 `json:"totalInfrastructureReports"`
	TotalEnvironmentReports    int64 `json:"totalEnvironmentReports"`
	TotalSafetyReports         int64 `json:"totalSafetyReports"`
	TotalTrafficReports        int64 `json:"totalTrafficReports"`
	TotalPublicFacilityReports int64 `json:"totalPublicFacilityReports"`
	TotalWasteReports          int64 `json:"totalWasteReports"`
	TotalWaterReports          int64 `json:"totalWaterReports"`
	TotalElectricityReports    int64 `json:"totalElectricityReports"`
	TotalHealthReports         int64 `json:"totalHealthReports"`
	TotalSocialReports         int64 `json:"totalSocialReports"`
	TotalEducationReports      int64 `json:"totalEducationReports"`
	TotalAdministrativeReports int64 `json:"totalAdministrativeReports"`
	TotalDisasterReports       int64 `json:"totalDisasterReports"`
	TotalOtherReports          int64 `json:"totalOtherReports"`
}

type Report struct {
	ID                         uint                        `json:"id"`
	ReportTitle                string                      `json:"reportTitle"`
	ReportType                 string                      `json:"reportType"`
	ReportDescription          string                      `json:"reportDescription"`
	ReportCreatedAt            int64                       `json:"reportCreatedAt"`
	ReportStatus               string                      `json:"reportStatus"`
	HasProgress                *bool                       `json:"hasProgress"`
	UserID                     uint                        `json:"userID"`
	UserName                   string                      `json:"userName"`
	FullName                   string                      `json:"fullName"`
	ProfilePicture             *string                     `json:"profilePicture"`
	Location                   ReportLocation              `json:"location"`
	Images                     ReportImage                 `json:"images"`
	TotalReactions             int64                       `json:"totalReactions"`
	TotalLikeReactions         *int64                      `json:"totalLikeReactions"`
	TotalDislikeReactions      *int64                      `json:"totalDislikeReactions"`
	TotalResolvedVotes         *int64                      `json:"totalResolvedVotes"`
	TotalOnProgressVotes       *int64                      `json:"totalOnProgressVotes"`
	TotalNotResolvedVotes      *int64                      `json:"totalNotResolvedVotes"`
	TotalVotes                 int64                       `json:"totalVotes"`
	IsLikedByCurrentUser       bool                        `json:"isLikedByCurrentUser"`
	IsDislikedByCurrentUser    bool                        `json:"isDislikedByCurrentUser"`
	ReportReactions            []ReactReportResponse       `json:"reportReactions"`
	ReportProgress             []GetProgressReportResponse `json:"reportProgress"`
	ReportVotes                []GetVoteReportResponse     `json:"reportVotes,omitempty"`
	IsResolvedByCurrentUser    bool                        `json:"isResolvedByCurrentUser"`
	IsOnProgressByCurrentUser  bool                        `json:"isOnProgressByCurrentUser"`
	IsNotResolvedByCurrentUser bool                        `json:"isNotResolvedByCurrentUser"`
	MajorityVote               *string                     `json:"majorityVote,omitempty"`
	LastUpdatedBy              *string                     `json:"lastUpdatedBy,omitempty"`
	LastUpdatedProgressAt      *int64                      `json:"lastUpdatedProgressAt,omitempty"`
	ReportUpdatedAt            int64                       `json:"reportUpdatedAt"`
}

type Distance struct {
	Distance string `json:"distance"`
	Lat      string `json:"lat"`
	Lng      string `json:"lng"`
}

type ReportLocation struct {
	DetailLocation string  `json:"detailLocation"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	DisplayName    *string `json:"displayName"`
	MapZoom        *int    `json:"mapZoom"`
	AddressType    *string `json:"addressType"`
	Country        *string `json:"country"`
	CountryCode    *string `json:"countryCode"`
	Region         *string `json:"region"`
	Road           *string `json:"road"`
	PostCode       *string `json:"postCode"`
	County         *string `json:"county"`
	State          *string `json:"state"`
	Village        *string `json:"village"`
	Suburb         *string `json:"suburb"`
	Geometry       *string `json:"geometry"`
}

type ReportImage struct {
	Image1URL *string `json:"image1URL"`
	Image2URL *string `json:"image2URL"`
	Image3URL *string `json:"image3URL"`
	Image4URL *string `json:"image4URL"`
	Image5URL *string `json:"image5URL"`
}

type CommentMedia struct {
	URL    string `json:"url"`
	Type   string `json:"type"`
	Width  *uint  `json:"width,omitempty"`
	Height *uint  `json:"height,omitempty"`
}

type Comment struct {
	CommentID       string                `json:"commentID"`
	ReportID        uint                  `json:"reportID"`
	UserInformation userDTO.UserProfile   `json:"userInformation"`
	Content         *string               `json:"content,omitempty"`
	Media           *CommentMedia         `json:"media,omitempty"`
	Mentions        []userDTO.UserProfile `json:"mentions,omitempty"`
	ReplyTo         *userDTO.UserProfile  `json:"replyTo,omitempty"`
	ThreadRootID    *string               `json:"threadRootID,omitempty"`
	ParentCommentID *string               `json:"parentCommentID,omitempty"`
	TotalReplies    int64                 `json:"totalReplies"`
	CreatedAt       int64                 `json:"createdAt"`
	UpdatedAt       *int64                `json:"updatedAt,omitempty"`
}

type CommentReply struct {
	CommentID       string                `json:"commentID"`
	ReportID        uint                  `json:"reportID"`
	UserInformation userDTO.UserProfile   `json:"userInformation"`
	Content         *string               `json:"content,omitempty"`
	Media           *CommentMedia         `json:"media,omitempty"`
	Mentions        []userDTO.UserProfile `json:"mentions,omitempty"`
	ReplyTo         *userDTO.UserProfile  `json:"replyTo,omitempty"`
	ThreadRootID    *string               `json:"threadRootID"`
	ParentCommentID *string               `json:"parentCommentID"`
	CreatedAt       int64                 `json:"createdAt"`
	UpdatedAt       *int64                `json:"updatedAt,omitempty"`
}