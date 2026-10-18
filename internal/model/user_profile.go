package model

type UserProfile struct {
    ID             uint    `json:"id" gorm:"primaryKey"`
    UserID         uint    `json:"userId" gorm:"unique;not null"`
    Bio            *string `json:"bio,omitempty" gorm:"type:text"`
    ProfilePicture *string `json:"profilePicture,omitempty" gorm:"size:255"`
    Gender         *string `json:"gender,omitempty" gorm:"size:20"`
    Birthday       *string `json:"birthday,omitempty" gorm:"type:date"`
}