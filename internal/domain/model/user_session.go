package model

type UserSession struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"not null"`
	User           User   `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	RefreshTokenID string `gorm:"type:varchar(64);not null;uniqueIndex"`
	HashedRefreshToken string `gorm:"type:varchar(256);not null"`
	IPAddress      string `gorm:"type:varchar(100);not null"`
	UserAgent      string `gorm:"type:text;not null"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
	IsActive       bool   `gorm:"default:true"`
	ExpiresAt      int64  `gorm:"not null"`
}
