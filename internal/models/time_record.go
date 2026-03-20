package models

import "time"

type TimeRecord struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	EventType  string    `json:"event_type"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	PhotoPath  string    `json:"photo_path"`
	PhotoURL   string    `json:"photo_url"`
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type TimeRecordListResponse struct {
	Data  []TimeRecord `json:"data"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

type TimeRecordTodayResponse struct {
	Data []TimeRecord `json:"data"`
}
