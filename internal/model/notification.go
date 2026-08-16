package model

type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "INFO"
	NotificationTypeWarning NotificationType = "WARNING"
	NotificationTypeError   NotificationType = "ERROR"
)

type NotificationCategory string

const (
	GeneralNotificationCategory NotificationCategory = "GENERAL"
	ReportNotificationCategory  NotificationCategory = "REPORT"
	UserNotificationCategory    NotificationCategory = "USER"
)

type EntityType string

const (
	EntityTypeReport EntityType = "REPORT"
	EntityTypeUser   EntityType = "USER"
	EntityTypeComment EntityType = "COMMENT"
)

type Notification struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"not null"`
	User           User   `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Title          string `gorm:"size:255;not null"`
	Description    string `gorm:"size:1000;not null"`
	Type           NotificationType           `gorm:"size:50;not null"`
	Category       NotificationCategory       `gorm:"size:50;not null"`
	IsRead         *bool  `gorm:"default:false"`
	ReadAt         *int64 `gorm:"default:null"`
	EntityID       *string `gorm:"default:null"`
	EntityType     *EntityType `gorm:"size:50;default:null"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
	DeletedAt      *int64 `gorm:"default:null"`
	IsDeleted      *bool  `gorm:"default:false"`
}
