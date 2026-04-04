package dto

type CreateReportRequest struct {
	ReportTitle       string  `json:"reportTitle" validate:"required";max=200"`
	ReportType        string  `json:"reportType" validate:"required,oneof=INFRASTRUCTURE ENVIRONMENT SAFETY TRAFFIC PUBLIC_FACILITY WASTE WATER ELECTRICITY HEALTH SOCIAL EDUCATION ADMINISTRATIVE DISASTER OTHER"`
	ReportDescription string  `json:"reportDescription" validate:"required"`
	DetailLocation    string  `json:"detailLocation" validate:"required"`
	HasProgress       *bool   `json:"hasProgress" validate:"omitempty"`
	MapZoom           *int    `json:"mapZoom" validate:"omitempty,min=0,max=21"`
	Latitude          float64 `json:"latitude" validate:"required"`
	Longitude         float64 `json:"longitude" validate:"required"`
	DisplayName       *string `json:"displayName" validate:"omitempty,max=255"`
	AddressType       *string `json:"addressType" validate:"omitempty,max=100"`
	Country           *string `json:"country" validate:"omitempty,max=100"`
	CountryCode       *string `json:"countryCode" validate:"omitempty,max=10"`
	Region            *string `json:"region" validate:"omitempty,max=100"`
	PostCode          *string `json:"postCode" validate:"omitempty,max=20"`
	County            *string `json:"county" validate:"omitempty,max=200"`
	State             *string `json:"state" validate:"omitempty,max=200"`
	Road              *string `json:"road" validate:"omitempty,max=200"`
	Village           *string `json:"village" validate:"omitempty,max=200"`
	Suburb            *string `json:"suburb" validate:"omitempty,max=200"`
	Image1URL         *string `json:"image1Url" validate:"omitempty,max=255"`
	Image2URL         *string `json:"image2Url" validate:"omitempty,max=255"`
	Image3URL         *string `json:"image3Url" validate:"omitempty,max=255"`
	Image4URL         *string `json:"image4Url" validate:"omitempty,max=255"`
	Image5URL         *string `json:"image5Url" validate:"omitempty,max=255"`
}

type EditReportRequest struct {
	ReportTitle       string  `json:"reportTitle" validate:"required";max=200"`
	ReportType        string  `json:"reportType" validate:"required,oneof=INFRASTRUCTURE ENVIRONMENT SAFETY TRAFFIC PUBLIC_FACILITY WASTE WATER ELECTRICITY HEALTH SOCIAL EDUCATION ADMINISTRATIVE DISASTER OTHER"`
	ReportDescription string  `json:"reportDescription" validate:"required"`
	DetailLocation    string  `json:"detailLocation" validate:"required"`
	HasProgress       *bool   `json:"hasProgress" validate:"omitempty"`
	Latitude          float64 `json:"latitude" validate:"required"`
	MapZoom           *int    `json:"mapZoom" validate:"omitempty,min=0,max=21"`
	Longitude         float64 `json:"longitude" validate:"required"`
	DisplayName       *string `json:"displayName" validate:"omitempty,max=255"`
	AddressType       *string `json:"addressType" validate:"omitempty,max=100"`
	Country           *string `json:"country" validate:"omitempty,max=100"`
	CountryCode       *string `json:"countryCode" validate:"omitempty,max=10"`
	Region            *string `json:"region" validate:"omitempty,max=100"`
	PostCode          *string `json:"postCode" validate:"omitempty,max=20"`
	County            *string `json:"county" validate:"omitempty,max=200"`
	State             *string `json:"state" validate:"omitempty,max=200"`
	Road              *string `json:"road" validate:"omitempty,max=200"`
	Village           *string `json:"village" validate:"omitempty,max=200"`
	Suburb            *string `json:"suburb" validate:"omitempty,max=200"`
	Image1URL         *string `json:"image1Url" validate:"omitempty,max=255"`
	Image2URL         *string `json:"image2Url" validate:"omitempty,max=255"`
	Image3URL         *string `json:"image3Url" validate:"omitempty,max=255"`
	Image4URL         *string `json:"image4Url" validate:"omitempty,max=255"`
	Image5URL         *string `json:"image5Url" validate:"omitempty,max=255"`
}

type ReactionReportRequest struct {
	ReactionType string `json:"reactionType" validate:"required,oneof=LIKE DISLIKE"`
}

type VoteReportRequest struct {
	VoteType string `json:"voteType" validate:"required,oneof=RESOLVED ON_PROGRESS NOT_RESOLVED"`
}

type UploadProgressReportRequest struct {
	Status      string  `json:"status" validate:"required,oneof=RESOLVED NOT_RESOLVED ON_PROGRESS"`
	Notes       string  `json:"notes" validate:"omitempty"`
	Attachment1 *string `json:"attachment1" validate:"omitempty,max=255"`
	Attachment2 *string `json:"attachment2" validate:"omitempty,max=255"`
}

type CreateReportCommentRequest struct {
	Content         *string `json:"content" validate:"omitempty,max=1000"`
	MediaURL        *string `json:"mediaURL" validate:"omitempty,max=255"`
	MediaType       *string `json:"mediaType" validate:"omitempty,oneof=IMAGE GIF VIDEO"`
	MediaWidth      *uint   `json:"mediaWidth" validate:"omitempty,min=1"`
	MediaHeight     *uint   `json:"mediaHeight" validate:"omitempty,min=1"`
	Mentions        []uint  `json:"mentions" validate:"omitempty,dive,gt=0"`
	ThreadRootID    *string `json:"threadRootID" validate:"omitempty,len=24"`
	ParentCommentID *string `json:"parentCommentID" validate:"omitempty,len=24"`
}