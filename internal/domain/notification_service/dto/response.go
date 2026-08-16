package dto

type GetNotificationsResponse struct {
	Notifications   []*Notification `json:"notifications"`
}