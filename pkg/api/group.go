package api

import "net/http"

type APIGroup interface {
	ListAPIs() []*API
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

type API struct {
	Pattern string
	Handler HandlerFunc
}

func NewSimpleAPI(pattern string, handler HandlerFunc) *API {
	return &API{
		Pattern: pattern,
		Handler: handler,
	}
}

func (a *API) ListAPIs() []*API {
	return []*API{a}
}
