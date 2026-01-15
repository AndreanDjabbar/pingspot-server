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
	POTENTIALLY_RESOLVED ReportStatus = "POTENTIALLY_RESOLVED"
	NOT_RESOLVED ReportStatus = "NOT_RESOLVED"
	ON_PROGRESS  ReportStatus = "ON_PROGRESS"
	WAITING      ReportStatus = "WAITING"
	EXPIRED	 	 ReportStatus = "EXPIRED"

	System      LastUpdatedBy = "SYSTEM"
	Owner	   	LastUpdatedBy = "OWNER"
)

type Report struct {
	ID                uint              `gorm:"primaryKey"`
	UserID            uint              `gorm:"not null"`
	User              User              `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ReportTitle       string            `gorm:"size:255;not null"`
	ReportType        ReportType        `gorm:"type:varchar(30);not null"`
	ReportDescription string            `gorm:"type:text;not null"`
	CreatedAt         int64             `gorm:"autoCreateTime"`
	UpdatedAt int64 `gorm:"autoUpdateTime:false"`
	PotentiallyResolvedAt *int64            `gorm:"default:null"`
	ReportStatus      ReportStatus      `gorm:"type:varchar(50);default:'WAITING'"`
	HasProgress       *bool             `gorm:"default:true"`
	LastUpdatedBy 	LastUpdatedBy `gorm:"type:varchar(50);default:NULL"`
	LastUpdatedProgressAt *int64            `gorm:"default:null"`
	AdminOverride   *bool             `gorm:"default:false"`
	IsDeleted         *bool             `gorm:"default:false"`
	DeletedAt 		*int64            `gorm:"default:null"`
	SearchVector 		string `gorm:"column:search_vector;->"`
	ReportLocation    *ReportLocation   `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ReportImages      *ReportImage      `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ReportReactions   *[]ReportReaction `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ReportProgress    *[]ReportProgress `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ReportVotes       *[]ReportVote     `gorm:"foreignKey:ReportID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
