package dto

import "time"

type PasteRequest struct {
	Content  string    `json:"content" validate:"required,min=5"`
	Privacy  string    `json:"privacy" validate:"required,oneof=private protected public"`
	Password string    `json:"password" validate:"password_required_if_protected"`
	ExpireAt time.Time `json:"expire_at"`
}
