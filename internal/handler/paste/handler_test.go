package paste

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/StewardMcCormick/Paste_Bin/internal/domain"
	"github.com/StewardMcCormick/Paste_Bin/internal/dto"
	errs "github.com/StewardMcCormick/Paste_Bin/internal/error"
	"github.com/StewardMcCormick/Paste_Bin/internal/handler/paste/mocks"
	"github.com/go-chi/chi/v5"
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

	s.handler = NewHandlers(s.useCase)
}

func (s *HandlerTestSuite) Test_Create_Success() {
	now := time.Now()
	afterWeek := now.Add(24 * 7 * time.Hour)
	expected := &dto.PasteResponse{
		Id:        1,
		Views:     0,
		Privacy:   string(domain.PublicPolicy),
		CreatedAt: now,
		ExpireAt:  afterWeek,
		Content:   "content",
		Hash:      "hash",
	}

	expectedLocation := "/api/v1/paste/" + expected.Hash

	testPasteRequest := &dto.PasteRequest{
		Content:  expected.Content,
		Privacy:  expected.Privacy,
		Password: "pass",
	}

	s.useCase.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(expected, nil).
		Once()

	body, err := json.Marshal(testPasteRequest)
	if err != nil {
		panic(err)
	}

	req := httptest.NewRequest("POST", "/api/v1/paste", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.handler.Create(w, req)

	resultBody, err := io.ReadAll(w.Body)
	if err != nil {
		panic(err)
	}
	result := &dto.PasteResponse{}
	err = json.NewDecoder(bytes.NewReader(resultBody)).Decode(result)
	if err != nil {
		panic(err)
	}

	s.Equal(result.Id, expected.Id)
	s.Equal(result.Views, expected.Views)
	s.Equal(result.Privacy, expected.Privacy)
	s.True(result.CreatedAt.Equal(expected.CreatedAt))
	s.True(result.ExpireAt.Equal(expected.ExpireAt))
	s.Equal(result.Content, expected.Content)
	s.Equal(expectedLocation, w.Header().Get("Location"))
}

func (s *HandlerTestSuite) Test_Create_Error() {
	cases := []struct {
		name           string
		setup          func()
		value          interface{}
		expectedStatus int
	}{
		{
			"Invalid JSON",
			func() {

			},
			"{ invalid JSON }",
			http.StatusBadRequest,
		},
		{
			"Internal error",
			func() {
				s.useCase.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil, errs.InternalError).
					Once()
			},
			&dto.PasteRequest{},
			http.StatusInternalServerError,
		},
		{
			"Bad Request error",
			func() {
				s.useCase.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil, errs.ValidationProcessError).
					Once()
			},
			&dto.PasteRequest{},
			http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			body, err := json.Marshal(tc.value)
			if err != nil {
				panic(err)
			}

			req := httptest.NewRequest("POST", "/api/v1/paste", bytes.NewReader(body))
			w := httptest.NewRecorder()

			s.handler.Create(w, req)

			s.Equal(tc.expectedStatus, w.Result().StatusCode)
		})
	}
}

func (s *HandlerTestSuite) Test_Get_Success() {
	now := time.Now()
	afterWeek := now.Add(24 * 7 * time.Hour)
	expectedPaste := &dto.PasteResponse{
		Id:        1,
		Views:     0,
		Privacy:   string(domain.PublicPolicy),
		CreatedAt: now,
		ExpireAt:  afterWeek,
		Content:   "content",
		Hash:      "hash",
	}
	reqArg := &dto.GetPasteRequest{Password: "pass"}
	s.useCase.EXPECT().
		GetByHash(mock.Anything,
			mock.MatchedBy(func(getRequest dto.GetPasteRequest) bool {
				return getRequest.Password == reqArg.Password
			}),
			mock.MatchedBy(func(h string) bool {
				return h == expectedPaste.Hash
			}),
		).
		Return(expectedPaste, nil).
		Once()

	reqBody, err := json.Marshal(reqArg)
	if err != nil {
		panic(err)
	}

	req := httptest.NewRequest("GET", "/api/v1/paste/"+expectedPaste.Hash, bytes.NewReader(reqBody))
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("pasteHash", expectedPaste.Hash)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	w := httptest.NewRecorder()

	s.handler.GetPaste(w, req)

	res := &dto.PasteResponse{}
	err = json.NewDecoder(w.Body).Decode(res)
	if err != nil {
		panic(err)
	}

	s.Equal(http.StatusOK, w.Code)
	s.Equal(expectedPaste.Id, res.Id)
	s.Equal(expectedPaste.Views, res.Views)
	s.Equal(expectedPaste.Content, res.Content)
	s.Equal(expectedPaste.Privacy, res.Privacy)
	s.True(expectedPaste.CreatedAt.Equal(res.CreatedAt))
	s.True(expectedPaste.ExpireAt.Equal(res.ExpireAt))
}

func (s *HandlerTestSuite) Test_Get_Error() {
	cases := []struct {
		name           string
		setup          func()
		value          interface{}
		expectedStatus int
	}{
		{
			"Request with invalid JSON",
			func() {
			},
			"{ invalid JSON }",
			http.StatusBadRequest,
		},
		{
			"Paste Not Found",
			func() {
				s.useCase.EXPECT().
					GetByHash(mock.Anything, mock.MatchedBy(func(req dto.GetPasteRequest) bool {
						return req.Password == "pass"
					},
					), mock.MatchedBy(func(hash string) bool {
						return hash == "hash"
					})).
					Return(nil, errs.PasteNotFound)
			},
			&dto.GetPasteRequest{Password: "pass"},
			http.StatusNotFound,
		},
		{
			"Forbidden",
			func() {
				s.useCase.EXPECT().
					GetByHash(mock.Anything, mock.MatchedBy(func(req dto.GetPasteRequest) bool {
						return req.Password == "pass"
					},
					), mock.MatchedBy(func(hash string) bool {
						return hash == "hash"
					})).
					Return(nil, errs.Forbidden)
			},
			&dto.GetPasteRequest{Password: "pass"},
			http.StatusForbidden,
		},
		{
			"Unauthorized",
			func() {
				s.useCase.EXPECT().
					GetByHash(mock.Anything, mock.MatchedBy(func(req dto.GetPasteRequest) bool {
						return req.Password == "pass"
					},
					), mock.MatchedBy(func(hash string) bool {
						return hash == "hash"
					})).
					Return(nil, errs.Unauthorized)
			},
			&dto.GetPasteRequest{Password: "pass"},
			http.StatusUnauthorized,
		},
		{
			"Internal Error",
			func() {
				s.useCase.EXPECT().
					GetByHash(mock.Anything, mock.MatchedBy(func(req dto.GetPasteRequest) bool {
						return req.Password == "pass"
					},
					), mock.MatchedBy(func(hash string) bool {
						return hash == "hash"
					})).
					Return(nil, errs.InternalError)
			},
			&dto.GetPasteRequest{Password: "pass"},
			http.StatusInternalServerError,
		},
		{
			"Unknown Error",
			func() {
				s.useCase.EXPECT().
					GetByHash(mock.Anything, mock.MatchedBy(func(req dto.GetPasteRequest) bool {
						return req.Password == "pass"
					},
					), mock.MatchedBy(func(hash string) bool {
						return hash == "hash"
					})).
					Return(nil, errors.New("unknown error"))
			},
			&dto.GetPasteRequest{Password: "pass"},
			http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.setup()

			reqBody, err := json.Marshal(tc.value)
			if err != nil {
				panic(err)
			}

			req := httptest.NewRequest("GET", "/api/v1/paste/hash", bytes.NewReader(reqBody))
			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("pasteHash", "hash")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			w := httptest.NewRecorder()

			s.handler.GetPaste(w, req)

			s.Equal(tc.expectedStatus, w.Code)
		})
	}
}
