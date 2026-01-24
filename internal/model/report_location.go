package model

type ReportLocation struct {
	ID             uint    `gorm:"primaryKey"`
	ReportID       uint    `gorm:"unique;not null"`
	Report         Report  `gorm:"foreignKey:ReportID;references:ID"`
	DetailLocation string  `gorm:"type:text;not null"`
	Latitude       float64 `gorm:"not null"`
	Longitude      float64 `gorm:"not null"`
	Geometry       string  `gorm:"type:geometry(Point, 4326);not null;default:ST_SetSRID(ST_MakePoint(0,0), 4326)"`
	DisplayName    *string `gorm:"type:text"`
	MapZoom		   *int    `gorm:"type:int"`
	AddressType    *string `gorm:"size:100"`
	Country        *string `gorm:"size:100"`
	CountryCode    *string `gorm:"size:10"`
	Region         *string `gorm:"size:100"`
	Road           *string `gorm:"size:200"`
	PostCode       *string `gorm:"size:20"`
	County         *string `gorm:"size:200"`
	State          *string `gorm:"size:200"`
	Village        *string `gorm:"size:200"`
	Suburb         *string `gorm:"size:200"`
}