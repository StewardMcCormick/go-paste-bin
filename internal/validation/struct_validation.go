package validation

import (
	"net/http"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/go-playground/validator/v10"
)

var (
	RequiredFieldMessage      = "This field is required"
	MinLengthMessage          = "Minimum length - "
	MaxLengthMessage          = "Maximum length - "
	RequiredPasswordIfMessage = "This field must be longer then 7 and shorten then 21 if Protected"
	OneOfMessage              = "Value of this field should be one of: "
)

type dtoTypes interface {
	*dto.UserRequest | *dto.PasteRequest
}

type appValidator[T dtoTypes] struct {
	valid *validator.Validate
}

func NewValidator[T dtoTypes](valid *validator.Validate) *appValidator[T] {
	err := valid.RegisterValidation("password_required_if_protected", passwordRequiredIfProtectedRule)
	if err != nil {
		panic(err)
	}

	return &appValidator[T]{valid: valid}
}

func (uv *appValidator[T]) Validate(request T) error {
	if err := uv.valid.Struct(request); err != nil {
		return uv.mapValidErrorToCustomError(err.(validator.ValidationErrors))
	}

	return nil
}

func (uv *appValidator[T]) mapValidErrorToCustomError(err validator.ValidationErrors) error {
	ve := errs.ValidationError{
		Message: errs.ValidationProcessError.Error(),
		Status:  http.StatusBadRequest,
		Errors:  make([]errs.ValidationFieldError, len(err)),
	}

	for i, fe := range err {
		ve.Errors[i] = errs.ValidationFieldError{
			Field:   fe.Field(),
			Message: uv.convertTagToMessage(fe),
			Value:   fe.Value(),
		}
	}

	return ve
}

func (uv *appValidator[T]) convertTagToMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return RequiredFieldMessage
	case "min":
		return MinLengthMessage + fe.Param()
	case "max":
		return MaxLengthMessage + fe.Param()
	case "password_required_if_protected":
		return RequiredPasswordIfMessage + fe.Param()
	case "oneof":
		return OneOfMessage + fe.Param()
	default:
		return fe.Error()
	}
}

func passwordRequiredIfProtectedRule(fl validator.FieldLevel) bool {
	req := fl.Parent().Interface().(dto.PasteRequest)

	if req.Privacy != string(domain.ProtectedPolicy) {
		return true
	}

	password := fl.Field().String()
	return 8 <= len(password) && len(password) <= 20
}
