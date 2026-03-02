package domain

import (
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
)

type PrivacyPolicy string
type PasteContent string

var (
	PrivatePolicy   PrivacyPolicy = "private"
	ProtectedPolicy PrivacyPolicy = "protected"
	PublicPolicy    PrivacyPolicy = "public"
)

type Paste struct {
	Id           int64
	UserId       int64
	Hash         string
	Views        int
	Privacy      PrivacyPolicy
	PasswordHash string
	CreatedAt    time.Time
	ExpireAt     time.Time
	Content      PasteContent
}

func (p *Paste) ToResponse() *dto.PasteResponse {
	return &dto.PasteResponse{
		Id:        p.Id,
		Views:     p.Views,
		Privacy:   string(p.Privacy),
		CreatedAt: p.CreatedAt,
		ExpireAt:  p.ExpireAt,
		Hash:      p.Hash,
		Content:   string(p.Content),
	}
}
