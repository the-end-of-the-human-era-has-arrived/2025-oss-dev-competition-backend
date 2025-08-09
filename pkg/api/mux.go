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

type RequestIDKey struct{}

func (m *APIServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	ctx := context.WithValue(r.Context(), RequestIDKey{}, requestID)
	r = r.WithContext(ctx)

	if _, p := m.mux.Handler(r); p == "" {
		http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
		return
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
			http.Error(w, apiError.Error(), apiError.StatusCode())
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
