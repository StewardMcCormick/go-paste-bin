package paste

import (
	"context"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	views "github.com/StewardMcCormick/Paste_Bin/internal/util/views_worker"
)

type Config struct {
	DefaultPasteExpiresTime time.Duration `yaml:"default_paste_expires_time" evn-default:"168h"`
}

type Repository interface {
	Create(context.Context, *domain.Paste) (*domain.Paste, error)
	GetByHash(ctx context.Context, hash string) (*domain.Paste, error)
	Update(ctx context.Context, paste *domain.Paste) (*domain.Paste, error)
}

type CreateRequestValidator interface {
	Validate(request *dto.PasteRequest) error
}

type UpdateRequestValidator interface {
	Validate(request *dto.UpdatePasteRequest) error
}

type Security interface {
	HashPassword(password string) (string, error)
	CompareHashAndPassword(hash string, pass string) bool
	GeneratePasteHash() (string, error)
}

type ViewWorker interface {
	SendEvent(ctx context.Context, event views.ViewEvent)
}

type UseCase struct {
	cfg                Config
	repo               Repository
	createRequestValid CreateRequestValidator
	updateRequestValid UpdateRequestValidator
	security           Security
	viewWorker         ViewWorker
}

func NewUseCase(
	cfg Config,
	repo Repository,
	createRequestValid CreateRequestValidator,
	updateRequestValid UpdateRequestValidator,
	security Security,
	viewWorker ViewWorker) *UseCase {
	return &UseCase{
		cfg:                cfg,
		repo:               repo,
		createRequestValid: createRequestValid,
		updateRequestValid: updateRequestValid,
		security:           security,
		viewWorker:         viewWorker,
	}
}
