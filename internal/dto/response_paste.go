package dto

import "time"

type PasteResponse struct {
	Id        int64     `json:"id"`
	Views     int       `json:"views,omitempty"`
	Privacy   string    `json:"privacy"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at,omitempty"`
	Content   string    `json:"content"`
}
