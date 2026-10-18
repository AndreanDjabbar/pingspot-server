package model

type ReportType string
type ReportStatus string
type LastUpdatedBy string

const (
	Infrastructure ReportType = "INFRASTRUCTURE"
	Environment    ReportType = "ENVIRONMENT"
	Safety         ReportType = "SAFETY"
	Traffic        ReportType = "TRAFFIC"
	PublicFacility ReportType = "PUBLIC_FACILITY"
	Waste          ReportType = "WASTE"
	Water          ReportType = "WATER"
	Electricity    ReportType = "ELECTRICITY"
	Health         ReportType = "HEALTH"
	Social         ReportType = "SOCIAL"
	Education      ReportType = "EDUCATION"
	Administrative ReportType = "ADMINISTRATIVE"
	Disaster       ReportType = "DISASTER"
	Other          ReportType = "OTHER"

	RESOLVED     ReportStatus = "RESOLVED"
	WAITING_CONFIRMATION ReportStatus = "WAITING_CONFIRMATION"
	ON_PROGRESS  ReportStatus = "ON_PROGRESS"
	WAITING      ReportStatus = "WAITING"
	EXPIRED	 	 ReportStatus = "EXPIRED"

	System      LastUpdatedBy = "SYSTEM"
	Owner	   	LastUpdatedBy = "OWNER"
)

type Report struct {
    ID                uint              `json:"id" gorm:"primaryKey"`
    UserID            uint              `json:"userId" gorm:"not null"`
    User              User              `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    ReportTitle       string            `json:"reportTitle" gorm:"size:255;not null"`
    ReportType        ReportType        `json:"reportType" gorm:"type:varchar(30);not null"`
    ReportDescription string            `json:"reportDescription" gorm:"type:text;not null"`
    CreatedAt         int64             `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt         int64             `json:"updatedAt" gorm:"autoUpdateTime:false"`
    PotentiallyResolvedAt *int64        `json:"potentiallyResolvedAt,omitempty" gorm:"default:null"`
    ReportStatus      ReportStatus      `json:"reportStatus" gorm:"type:varchar(50);default:'WAITING'"`
    HasProgress       *bool             `json:"hasProgress,omitempty" gorm:"default:true"`
    LastUpdatedBy     LastUpdatedBy     `json:"lastUpdatedBy,omitempty" gorm:"type:varchar(50);default:NULL"`
    LastUpdatedProgressAt *int64        `json:"lastUpdatedProgressAt,omitempty" gorm:"default:null"`
    AdminOverride     *bool             `json:"adminOverride,omitempty" gorm:"default:false"`
    IsDeleted         *bool             `json:"isDeleted,omitempty" gorm:"default:false"`
    DeletedAt         *int64            `json:"deletedAt,omitempty" gorm:"default:null"`
    SearchVector      string            `json:"-" gorm:"column:search_vector;->;-:migration"`
    ReportLocation    *ReportLocation   `json:"reportLocation,omitempty" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    ReportImages      *ReportImage      `json:"reportImages,omitempty" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    ReportReactions   *[]ReportReaction `json:"reportReactions,omitempty" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    ReportProgress    *[]ReportProgress `json:"reportProgress,omitempty" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    ReportVotes       *[]ReportVote     `json:"reportVotes,omitempty" gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
