package database

type VideoInformation struct {
	Id string `gorm:"column:Id;not null"`
	UUID string `gorm:"column:Uuid;not null"`
	Title string `gorm:"column:Title;not null"`
	Description string `gorm:"column:Description;not null"`
	CreationDate int64 `gorm:"column:CreationDate;not null"`
	VideoLength float64 `gorm:"column:VideoLength;not null"`
}

type VideoSession struct {
    LastTime             int64	`json:"LastTime" gorm:"column:last_time;not null"`
    VideoId              string	`json:"VideoId" gorm:"column:video_id;not null"`
    LastPosition         float64 `json:"LastPosition" gorm:"column:last_position;default:0"`
    TotalVerifiedSeconds float64 `json:"TotalVerifiedSeconds" gorm:"column:total_verified_seconds;default:0"`
}
