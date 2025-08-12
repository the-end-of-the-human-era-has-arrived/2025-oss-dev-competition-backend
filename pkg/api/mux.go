package api

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

const NotionUserIDCookie = "notionUserId"

type APIServeMux struct {
	mux *http.ServeMux
}

func newAPIServeMux() *APIServeMux {
	return &APIServeMux{
		mux: http.NewServeMux(),
	}
}

type (
	RequestIDKey struct{}
)

func (m *APIServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS (프론트: 3000)
	origin := r.Header.Get("Origin")
	fe := os.Getenv("FRONTEND_ORIGIN")
	if fe == "" {
		fe = "http://localhost:3000"
	}
	if origin == fe {
		w.Header().Set("Access-Control-Allow-Origin", fe)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, p := m.mux.Handler(r)
	if p == "" {
		ErrNotFound.WriteHTTPError(w)
		return
	}

	// 요청 ID
	requestID, err := uuid.NewRandom()
	if err != nil {
		ErrFailToCreateRequestID.WriteHTTPError(w)
		return
	}
	ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
	r = r.WithContext(ctx)

	// /api 보호: notionUserId 쿠키 → 세션 적재
	path := strings.Split(p, " ")[1]
	if strings.HasPrefix(path, "/api") {
		rc, err := r.Cookie(NotionUserIDCookie)
		if err != nil || rc.Value == "" {
			ErrInvalidSession.WriteHTTPError(w)
			return
		}
		// 쿠키값(문자열 NotionUserID)로 세션 로드
		if sess, ok := SessionStore.Get(rc.Value); ok && sess != nil {
			r = r.WithContext(context.WithValue(r.Context(), SessionKey{}, sess))
		} else {
			ErrInvalidSession.WriteHTTPError(w)
			return
		}
	}

	m.mux.ServeHTTP(w, r)
}

func (m *APIServeMux) RegistAPI(apis ...*API) {
	for _, a := range apis {
		m.mux.HandleFunc(a.Pattern, WithErrorHandler(a.Handler))
	}
}

func WithErrorHandler(handlerFn HandlerFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		apiError := &Error{}
		err := handlerFn(w, r)
		if errors.As(err, apiError) {
			apiError.WriteHTTPError(w)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
