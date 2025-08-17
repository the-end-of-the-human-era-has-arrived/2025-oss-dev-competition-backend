package controller

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/service"
)

type mindMapController struct {
	service *service.MindMapService
}

func NewMindMapController(service *service.MindMapService) *mindMapController {
	return &mindMapController{
		service: service,
	}
}

var _ api.APIGroup = (*mindMapController)(nil)

func (c *mindMapController) ListAPIs() []*api.API {
	return []*api.API{
		api.NewSimpleAPI("POST /api/users/{userID}/mindmap", c.createMindMap),
		api.NewSimpleAPI("GET /api/users/{userID}/mindmap", c.getMindMap),
		api.NewSimpleAPI("DELETE /api/users/{userID}/mindmap", c.deleteMindMap),
	}
}

func (c *mindMapController) createMindMap(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session, ok := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if !ok || session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }

	params := &struct {
		Nodes []*domain.KeywordNode `json:"nodes"`
		Edges []*domain.EdgeOfIndex `json:"edges"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(params); err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	for _, e := range params.Edges {
		if e.Idx1 < 0 || e.Idx2 < 0 || e.Idx1 >= len(params.Nodes) || e.Idx2 >= len(params.Nodes) {
			return api.NewError(http.StatusBadRequest, api.WithMessage("invalid index"))
		}
	}

	if err := c.service.BuildMindMap(r.Context(), userUID, params.Nodes, params.Edges); err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}
	return api.ResponseStatusCode(w, http.StatusCreated, "success to create mindmap")
}

func (c *mindMapController) getMindMap(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }

	graph := c.service.GetMindMapByUser(r.Context(), userUID)

	return api.ResponseJSON(w, graph)
}

func (c *mindMapController) deleteMindMap(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }

	if err := c.service.DeleteMindMapByUser(r.Context(), userUID); err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}
	return api.ResponseStatusCode(w, http.StatusOK, "success to delete mindmap")
}
