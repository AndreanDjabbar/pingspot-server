package model

type ReportLocation struct {
    ID             uint    `json:"id" gorm:"primaryKey"`
    ReportID       uint    `json:"reportId" gorm:"unique;not null"`
    Report         Report  `json:"report" gorm:"foreignKey:ReportID;references:ID"`
    DetailLocation string  `json:"detailLocation" gorm:"type:text;not null"`
    Latitude       float64 `json:"latitude" gorm:"not null"`
    Longitude      float64 `json:"longitude" gorm:"not null"`
    Geometry       string  `json:"geometry" gorm:"type:geometry(Point, 4326);not null;default:ST_SetSRID(ST_MakePoint(0,0), 4326)"`
    DisplayName    *string `json:"displayName,omitempty" gorm:"type:text"`
    MapZoom        *int    `json:"mapZoom,omitempty" gorm:"type:int"`
    AddressType    *string `json:"addressType,omitempty" gorm:"size:100"`
    Country        *string `json:"country,omitempty" gorm:"size:100"`
    CountryCode    *string `json:"countryCode,omitempty" gorm:"size:10"`
    Region         *string `json:"region,omitempty" gorm:"size:100"`
    Road           *string `json:"road,omitempty" gorm:"size:200"`
    PostCode       *string `json:"postCode,omitempty" gorm:"size:20"`
    County         *string `json:"county,omitempty" gorm:"size:200"`
    State          *string `json:"state,omitempty" gorm:"size:200"`
    Village        *string `json:"village,omitempty" gorm:"size:200"`
    Suburb         *string `json:"suburb,omitempty" gorm:"size:200"`
}