package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler/user/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	useCase *mocks.MockUseCase
	handler *httpHandlers
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func (s *HandlerTestSuite) SetupTest() {
	s.useCase = mocks.NewMockUseCase(s.T())
	s.handler = NewHandler(s.useCase)
}

func (s *HandlerTestSuite) Test_Login_Success() {
	now := time.Now()
	afterWeek := now.Add(24 * 7 * time.Hour)
	expectedKey := &dto.APIKeyResponse{Key: "key", ExpiresAt: afterWeek}
	testRequest := &dto.UserRequest{Username: "user", Password: "password"}

	s.useCase.EXPECT().
		Login(mock.Anything, testRequest).
		Return(expectedKey, nil).
		Once()

	reqBody, err := json.Marshal(testRequest)
	s.Require().NoError(err)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	s.handler.Login(w, req)

	respBody := &dto.APIKeyResponse{}
	err = json.NewDecoder(w.Result().Body).Decode(respBody)
	s.Require().NoError(err)

	s.Equal(http.StatusOK, w.Code)
	s.Equal(expectedKey.Key, respBody.Key)
	s.True(expectedKey.ExpiresAt.Equal(respBody.ExpiresAt))
}

func (s *HandlerTestSuite) Test_Login_Error() {
	cases := []struct {
		name         string
		setup        func()
		value        interface{}
		expectedCode int
	}{
		{
			"Invalid JSON",
			func() {
			},
			"{ Invalid JSON }",
			http.StatusBadRequest,
		},
		{
			"Not Found Error",
			func() {
				s.useCase.EXPECT().
					Login(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errs.ErrUserNotFound).
					Once()
			},
			&dto.UserRequest{
				Username: "user",
				Password: "password",
			},
			http.StatusNotFound,
		},
		{
			"ErrUnauthorized",
			func() {
				s.useCase.EXPECT().
					Login(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errs.ErrUnauthorized).
					Once()
			},
			&dto.UserRequest{
				Username: "user",
				Password: "password",
			},
			http.StatusUnauthorized,
		},
		{
			"Internal Error",
			func() {
				s.useCase.EXPECT().
					Login(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errs.ErrInternal).
					Once()
			},
			&dto.UserRequest{
				Username: "user",
				Password: "password",
			},
			http.StatusInternalServerError,
		},
		{
			"Validation Error",
			func() {
				s.useCase.EXPECT().
					Login(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errs.ValidationError{}).
					Once()
			},
			&dto.UserRequest{
				Username: "user",
				Password: "password",
			},
			http.StatusBadRequest,
		},
		{
			"Unknown Error",
			func() {
				s.useCase.EXPECT().
					Login(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errors.New("unknown error")).
					Once()
			},
			&dto.UserRequest{
				Username: "user",
				Password: "password",
			},
			http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			reqBody, err := json.Marshal(tc.value)
			s.Require().NoError(err)

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			s.handler.Login(w, req)

			s.Equal(tc.expectedCode, w.Code)
		})
	}
}

func (s *HandlerTestSuite) Test_Registration_Success() {
	now := time.Now()
	afterWeek := now.Add(24 * 7 * time.Hour)
	testRequest := &dto.UserRequest{Username: "user", Password: "password"}
	expectedUser := &dto.UserResponse{
		Id:        0,
		Username:  "user",
		APIKey:    dto.APIKeyResponse{Key: "key", ExpiresAt: afterWeek},
		CreatedAt: now,
	}

	s.useCase.EXPECT().
		Registration(mock.Anything, mock.MatchedBy(func(user *dto.UserRequest) bool {
			return user.Username == testRequest.Username && user.Password == testRequest.Password
		})).
		Return(expectedUser, nil).
		Once()

	reqBody, err := json.Marshal(testRequest)
	s.Require().NoError(err)

	req := httptest.NewRequest("POST", "/registartion", bytes.NewReader(reqBody))
	w := httptest.NewRecorder()

	s.handler.Registration(w, req)

	resp := &dto.UserResponse{}
	err = json.NewDecoder(w.Result().Body).Decode(resp)
	s.Require().NoError(err)

	s.Equal(http.StatusCreated, w.Code)
	s.Equal(expectedUser.Id, resp.Id)
	s.Equal(expectedUser.Username, resp.Username)
	s.Equal(expectedUser.APIKey.Key, resp.APIKey.Key)
	s.True(expectedUser.APIKey.ExpiresAt.Equal(resp.APIKey.ExpiresAt))
	s.True(expectedUser.CreatedAt.Equal(resp.CreatedAt))
}

func (s *HandlerTestSuite) Test_Registration_Error() {
	cases := []struct {
		name         string
		setup        func()
		value        interface{}
		expectedCode int
	}{
		{
			"Invalid JSON",
			func() {
			},
			"{ Invalid JSON }",
			http.StatusBadRequest,
		},
		{
			"User Already Exist Error",
			func() {
				s.useCase.EXPECT().
					Registration(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errs.ErrUserAlreadyExists).
					Once()
			},
			&dto.UserRequest{Username: "user", Password: "password"},
			http.StatusConflict,
		},
		{
			"Validation Error",
			func() {
				s.useCase.EXPECT().
					Registration(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errs.ValidationError{}).
					Once()
			},
			&dto.UserRequest{Username: "user", Password: "password"},
			http.StatusBadRequest,
		},
		{
			"Internal Error",
			func() {
				s.useCase.EXPECT().
					Registration(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errs.ErrInternal).
					Once()
			},
			&dto.UserRequest{Username: "user", Password: "password"},
			http.StatusInternalServerError,
		},
		{
			"Unknown Error",
			func() {
				s.useCase.EXPECT().
					Registration(mock.Anything, mock.MatchedBy(func(req *dto.UserRequest) bool {
						return req.Username == "user" && req.Password == "password"
					})).
					Return(nil, errors.New("unknown error")).
					Once()
			},
			&dto.UserRequest{Username: "user", Password: "password"},
			http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			body, err := json.Marshal(tc.value)
			s.Require().NoError(err)

			req := httptest.NewRequest("POST", "/registration", bytes.NewReader(body))
			w := httptest.NewRecorder()

			s.handler.Registration(w, req)

			s.Equal(tc.expectedCode, w.Code)
		})
	}
}
