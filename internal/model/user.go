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
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	Username   string    `gorm:"size:30;unique;not null"`
	Email      string    `gorm:"size:100;unique;not null"`
	Password   *string   `gorm:"size:255"`
	FullName   string    `gorm:"size:100;not null"`
	Provider   Provider  `gorm:"type:varchar(20);default:EMAIL;not null"`
	IsVerified bool      `gorm:"default:false;not null"`
	ProviderID *string   `gorm:"size:100"`
	Profile	UserProfile `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
	SearchVector string    `gorm:"column:search_vector;->"`
	
	Reports    []Report  `gorm:"foreignKey:UserID"`
}