package paste

import "github.com/StewardMcCormick/Paste_Bin/internal/adapter/postgres"

type Repository struct {
	Pool postgres.DBTX
}
