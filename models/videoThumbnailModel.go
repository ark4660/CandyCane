package models

type Video struct {
    Description  string `json:"description"`
    ThumbnailURL string `json:"thumbnail_url"`
    Title        string `json:"title"`
    VideoID      string `json:"video_id"`
}
