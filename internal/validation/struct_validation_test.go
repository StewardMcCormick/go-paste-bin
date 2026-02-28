package validation

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateUser(t *testing.T) {
	valid := NewValidator[*dto.UserRequest](validator.New(validator.WithRequiredStructEnabled()))

	cases := []struct {
		name     string
		value    *dto.UserRequest
		expected *errs.ValidationError
	}{
		{
			name:     "Correct user",
			value:    &dto.UserRequest{Username: "Correct_User", Password: "password"},
			expected: nil,
		},
		{
			"User with empty fields",
			&dto.UserRequest{Username: "", Password: ""},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Username", RequiredFieldMessage, ""},
					{"Password", RequiredFieldMessage, ""},
				},
			},
		},
		{
			"User with too shorten fields",
			&dto.UserRequest{Username: "Us", Password: "pass"},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Username", MinLengthMessage, "Us"},
					{"Password", MinLengthMessage, "pass"},
				},
			},
		},
		{
			"User with too longer fields",
			&dto.UserRequest{Username: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Username", MaxLengthMessage, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
					{"Password", MaxLengthMessage, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := valid.Validate(tc.value)

			if tc.expected == nil {
				assert.Nil(t, got)
				return
			}

			var err errs.ValidationError
			errors.As(got, &err)

			assert.Equal(t, tc.expected.Message, err.Message)
			assert.Equal(t, tc.expected.Status, err.Status)

			for i, fe := range err.Errors {
				assert.Equal(t, tc.expected.Errors[i].Field, fe.Field)
				assert.True(t, strings.HasPrefix(fe.Message, tc.expected.Errors[i].Message))
				assert.Equal(t, tc.expected.Errors[i].Value, fe.Value)
			}
		})
	}
}

func TestValidator_ValidatePaste(t *testing.T) {
	valid := NewValidator[*dto.PasteRequest](validator.New(validator.WithRequiredStructEnabled()))

	cases := []struct {
		name     string
		value    *dto.PasteRequest
		expected *errs.ValidationError
	}{
		{
			"Valid value",
			&dto.PasteRequest{
				Content:  "valid content",
				Privacy:  string(domain.PublicPolicy),
				Password: "valid_password",
			},
			nil,
		},
		{
			"Paste with too short fields",
			&dto.PasteRequest{
				Content:  "ic",
				Privacy:  string(domain.PublicPolicy),
				Password: "pass",
			},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Content", MinLengthMessage, "ic"},
					{"Password", MinLengthMessage, "pass"},
				},
			},
		},
		{
			"Paste with empty required fields",
			&dto.PasteRequest{
				Content:  "",
				Privacy:  "public",
				Password: "",
			},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Content", RequiredFieldMessage, ""},
				},
			},
		},
		{
			"Paste with invalid Privacy type",
			&dto.PasteRequest{
				Content:  "Content",
				Privacy:  "wrong",
				Password: "password",
			},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Privacy", OneOfMessage, "wrong"},
				},
			},
		},
		{
			"Paste with Protected Privacy type and too short password",
			&dto.PasteRequest{
				Content:  "Content",
				Privacy:  string(domain.ProtectedPolicy),
				Password: "pass",
			},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Password", RequiredPasswordIfMessage, "pass"},
				},
			},
		},
		{
			"Paste with Protected Privacy type and too long password",
			&dto.PasteRequest{
				Content:  "Content",
				Privacy:  string(domain.ProtectedPolicy),
				Password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			&errs.ValidationError{
				Message: errs.ValidationProcessError.Error(),
				Status:  http.StatusBadRequest,
				Errors: []errs.ValidationFieldError{
					{"Password", RequiredPasswordIfMessage, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := valid.Validate(tc.value)

			if tc.expected == nil {
				assert.Nil(t, got)
				return
			}

			var err errs.ValidationError
			errors.As(got, &err)

			assert.Equal(t, tc.expected.Message, err.Message)
			assert.Equal(t, tc.expected.Status, err.Status)

			for i, fe := range err.Errors {
				assert.Equal(t, tc.expected.Errors[i].Field, fe.Field)
				assert.True(t, strings.HasPrefix(fe.Message, tc.expected.Errors[i].Message))
				assert.Equal(t, tc.expected.Errors[i].Value, fe.Value)
			}
		})
	}
}
