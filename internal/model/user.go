package model

import (
	"time"
)

type Provider string

const (
	ProviderEmail    Provider = "EMAIL"
	ProviderGoogle   Provider = "GOOGLE"
	ProviderFacebook Provider = "FACEBOOK"
)

type User struct {
    ID                         uint        `json:"id" gorm:"primaryKey;autoIncrement"`
    Username                   string      `json:"username" gorm:"size:30;unique;not null"`
    Email                      string      `json:"email" gorm:"size:100;unique;not null"`
    Password                   *string     `json:"-" gorm:"size:255"`
    FullName                   string      `json:"fullName" gorm:"size:100;not null"`
    Provider                   Provider    `json:"provider" gorm:"type:varchar(20);default:EMAIL;not null"`
    IsVerified                 bool        `json:"isVerified" gorm:"default:false;not null"`
    ProviderID                 *string     `json:"providerId,omitempty" gorm:"size:100"`
    Profile                    UserProfile `json:"profile" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    IsDefaultUsername          bool        `json:"isDefaultUsername" gorm:"default:true;not null"`
    IsDisableEmailNotification bool        `json:"isDisableEmailNotification" gorm:"default:false;not null"`
    CreatedAt                  time.Time   `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt                  time.Time   `json:"updatedAt" gorm:"autoUpdateTime"`
    SearchVector               string      `json:"-" gorm:"column:search_vector;->;-:migration"`
    Reports                    []Report    `json:"reports,omitempty" gorm:"foreignKey:UserID"`
}