package database

import (
    "gorm.io/gorm"
    "fmt"
    "encoding/json"
)

func GetSession(db *gorm.DB, UUID string) ([]byte ,error) {
	var sessions []VideoSession
	db.Model(&VideoSession{}).
    	Select("last_position, total_verified_seconds, video_id, last_time").
     	Where("uuid = ?", UUID).
      	Order("last_time DESC").
       	Find(&sessions)
	jsonData, _ := json.Marshal(sessions)
	fmt.Println(string(jsonData))
	return jsonData, nil
}
