package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type APIServeMux struct {
	mux *http.ServeMux
}

func NewAPIServeMux() *APIServeMux {
	return &APIServeMux{
		mux: http.NewServeMux(),
	}
}

type (
	RequestIDKey struct{}
	SessionKey   struct{}
)

func (m *APIServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch _, p := m.mux.Handler(r); p {
	case "":
		ErrNotFound.WriteHTTPError(w)
		return
	case "/auth/notion", "/auth/notion/callback":
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

	ctx := context.WithValue(r.Context(), SessionKey{}, session)

	requestID, err := uuid.NewRandom()
	if err != nil {
		ErrFailToCreateRequestID.WriteHTTPError(w)
		return
	}

	ctx = context.WithValue(ctx, RequestIDKey{}, requestID)
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
