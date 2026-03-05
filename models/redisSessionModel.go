package models

type RedisSessionModel struct {
	UUID string	`json:"Uuid"`
	WatchSessionId string `json:"WatchSessionId"`
	LastPosition float64 `json:"LastPosition"`
}
