package validation

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestUserValidator_Validate(t *testing.T) {
	valid := NewUserValidator(validator.New(validator.WithRequiredStructEnabled()))

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
				Message: errs.UserValidationError.Error(),
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
				Message: errs.UserValidationError.Error(),
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
				Message: errs.UserValidationError.Error(),
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
