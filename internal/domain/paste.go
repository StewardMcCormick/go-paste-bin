package domain

import "time"

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
