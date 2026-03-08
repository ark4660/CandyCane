package database

import (
    "gorm.io/gorm"
)

func Upload(db *gorm.DB, videoData *VideoInformation) error {
	if err := db.Create(videoData).Error; err != nil {
		return err
    }
    return nil
}
