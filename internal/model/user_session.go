package model

type UserSession struct {
    ID                 uint   `json:"id" gorm:"primaryKey"`
    UserID             uint   `json:"userId" gorm:"not null"`
    User               User   `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    RefreshTokenID     string `json:"refreshTokenId" gorm:"type:varchar(64);not null;uniqueIndex"`
    HashedRefreshToken string `json:"-" gorm:"type:varchar(256);not null"`
    IPAddress          string `json:"ipAddress" gorm:"type:varchar(100);not null"`
    UserAgent          string `json:"userAgent" gorm:"type:text;not null"`
    CreatedAt          int64  `json:"createdAt" gorm:"autoCreateTime"`
    IsActive           bool   `json:"isActive" gorm:"default:true"`
    ExpiresAt          int64  `json:"expiresAt" gorm:"not null"`
}