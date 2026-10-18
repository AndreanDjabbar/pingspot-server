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
    ID             uint                 `json:"id" gorm:"primaryKey"`
    UserID         uint                 `json:"userId" gorm:"not null"`
    User           User                 `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Title          string               `json:"title" gorm:"size:255;not null"`
    Description    string               `json:"description" gorm:"size:1000;not null"`
    Type           NotificationType     `json:"type" gorm:"size:50;not null"`
    Category       NotificationCategory `json:"category" gorm:"size:50;not null"`
    IsRead         *bool                `json:"isRead,omitempty" gorm:"default:false"`
    ReadAt         *int64               `json:"readAt,omitempty" gorm:"default:null"`
    EntityID       *string              `json:"entityId,omitempty" gorm:"default:null"`
    EntityType     *EntityType          `json:"entityType,omitempty" gorm:"size:50;default:null"`
    CreatedAt      int64                `json:"createdAt" gorm:"autoCreateTime"`
    DeletedAt      *int64               `json:"deletedAt,omitempty" gorm:"default:null"`
    IsDeleted      *bool                `json:"isDeleted,omitempty" gorm:"default:false"`
}