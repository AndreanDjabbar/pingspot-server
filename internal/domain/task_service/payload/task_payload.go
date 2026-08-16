package payload

import "pingspot/internal/model"

type UpdateProgressPayload struct {
	ReportID uint `json:"report_id"`
}

type CreateNotificationPayload struct {
	UserID      uint            `json:"user_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	EntityID    *string         `json:"entity_id,omitempty"`
	EntityType  model.EntityType         `json:"entity_type,omitempty"`
	Category    model.NotificationCategory `json:"category,omitempty"`
	Type        model.NotificationType     `json:"type,omitempty"`
}