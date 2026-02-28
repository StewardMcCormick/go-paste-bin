package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	cfgUtil "github.com/StewardMcCormick/Paste_Bin/config/cfg_util"
	appctx "github.com/StewardMcCormick/Paste_Bin/internal/util/app_context"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
)

type Util struct {
}

func NewUtil() *Util {
	return &Util{}
}

func (s *Util) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func (s *Util) HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))

	return hex.EncodeToString(hash[:])
}

func (s *Util) GenerateAPIKey(ctx context.Context) (keyPrefix string, key string, err error) {
	var prefix string
	switch ctx.Value(appctx.EnvKey) {
	case cfgUtil.DevelopmentEnv:
		prefix = "pb_test"
	default:
		prefix = "pb_live"
	}

	randPart, err := gonanoid.New(4)
	if err != nil {
		return "", "", err
	}
	resultKey := prefix + "_" + randPart + "_"

	randPart, err = gonanoid.New(12)
	if err != nil {
		return "", "", err
	}

	resultKey += randPart

	return prefix, resultKey, nil
}

func (s *Util) CompareHashAndPassword(hash string, pass string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass)); err != nil {
		return false
	}

	return true
}

func (s *Util) GeneratePasteHash() (string, error) {
	hash, err := gonanoid.New(20)
	if err != nil {
		return "", err
	}

	return hash, nil
}
