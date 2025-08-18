package controller

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/service"
)

type notionPageController struct {
	service *service.NotionPageService
}

func NewNotionPageController(service *service.NotionPageService) *notionPageController {
	return &notionPageController{
		service: service,
	}
}

var _ api.APIGroup = (*notionPageController)(nil)

func (c *notionPageController) ListAPIs() []*api.API {
	return []*api.API{
		api.NewSimpleAPI("POST /api/users/{userID}/notion", c.createNotionPage),
		api.NewSimpleAPI("GET /api/users/{userID}/notion", c.getAllNotionPages),
		api.NewSimpleAPI("GET /api/users/{userID}/notion/{notionPageID}", c.getNotionPage),
		api.NewSimpleAPI("PUT /api/users/{userID}/notion/{notionPageID}", c.updateNotionPage),
		api.NewSimpleAPI("DELETE /api/users/{userID}/notion/{notionPageID}", c.deleteNotionPage),
	}
}

func (c *notionPageController) createNotionPage(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")
	requestID := r.Context().Value(api.RequestIDKey{}).(uuid.UUID)
	log.Println("requestID:", requestID.String(), "userID:", userID)

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}
	defer r.Body.Close()

	log.Println("requestID:", requestID.String(), "body:", string(body))

	param := &domain.NotionPage{}
	if err := json.Unmarshal(body, &param); err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	param.UserID = userUID

	page, err := c.service.CreateNotionPage(r.Context(), param)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	return api.ResponseJSON(r.Context(), w, page)
}

func (c *notionPageController) getAllNotionPages(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }

	pages, err := c.service.GetAllNotionPagesByUser(r.Context(), userUID)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	return api.ResponseJSON(r.Context(), w, pages)
}

func (c *notionPageController) getNotionPage(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")
	notionPageID := r.PathValue("notionPageID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	notionPageUID, err := uuid.Parse(notionPageID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }

	page, err := c.service.GetNotionPageByID(r.Context(), notionPageUID)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	if page.UserID != userUID {
		return api.NewError(http.StatusForbidden, api.WithMessage("access denied"))
	}

	return api.ResponseJSON(r.Context(), w, page)
}

func (c *notionPageController) updateNotionPage(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")
	notionPageID := r.PathValue("notionPageID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	notionPageUID, err := uuid.Parse(notionPageID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }

	param := &domain.NotionPage{}
	if err := json.NewDecoder(r.Body).Decode(param); err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}
	defer r.Body.Close()

	param.ID = notionPageUID
	param.UserID = userUID

	page, err := c.service.UpdateNotionPage(r.Context(), param)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	return api.ResponseJSON(r.Context(), w, page)
}

func (c *notionPageController) deleteNotionPage(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")
	notionPageID := r.PathValue("notionPageID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	notionPageUID, err := uuid.Parse(notionPageID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	// session := r.Context().Value(api.SessionKey{}).(*api.Session)
	// if session.UserID != userUID {
	// 	return api.ErrInvalidSession
	// }

	page, err := c.service.DeleteNotionPageByID(r.Context(), notionPageUID)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	if page.UserID != userUID {
		return api.NewError(http.StatusForbidden, api.WithMessage("access denied"))
	}

	return api.ResponseStatusCode(r.Context(), w, http.StatusOK, "success to delete notion page")
}
