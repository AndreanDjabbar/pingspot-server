package dto

type Notification struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"userID"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsRead      bool   `json:"isRead"`
	ReadAt      *int64 `json:"readAt,omitempty"`
	CreatedAt   int64  `json:"createdAt"`
	EntityID   *string `json:"entityID,omitempty"`
	EntityType *string `json:"entityType,omitempty"`
}