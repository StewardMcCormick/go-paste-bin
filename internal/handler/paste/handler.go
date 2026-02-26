package paste

type UseCase interface {
}

type httpHandlers struct {
	useCase UseCase
}

func NewHandlers(useCase UseCase) *httpHandlers {
	return &httpHandlers{useCase: useCase}
}
