package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

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
	SessionKey   struct{}
)

func (m *APIServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // CORS 헤더 설정 (모든 요청에 적용)
    w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    // OPTIONS 요청 (preflight) 처리
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

	_, pattern := m.mux.Handler(r)
	path := extractPath(pattern)
	if path == "" {
		ErrNotFound.WriteHTTPError(w)
		return
	}

	requestID, err := uuid.NewRandom()
	if err != nil {
		ErrFailToCreateRequestID.WriteHTTPError(w)
		return
	}

	ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)

	if !strings.HasPrefix(path, "/api") {
		m.mux.ServeHTTP(w, r)
		return
	}

	cookie, err := r.Cookie("sessionID")
	if err != nil {
		NewError(http.StatusUnauthorized, WithError(err)).WriteHTTPError(w)
		return
	}

	session, ok := SessionStore.Get(cookie.Value)
	if !ok {
		ErrNoSession.WriteHTTPError(w)
		return
	}

	ctx = context.WithValue(ctx, SessionKey{}, session)
	r = r.WithContext(ctx)

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

func extractPath(pattern string) string {
	if i := strings.Index(pattern, " "); i != -1 {
		return pattern[i+1:]
	}
	return pattern
}
