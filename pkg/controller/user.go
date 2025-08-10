package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/service"
)

type userController struct {
	service *service.UserService
}

func NewUserController(service *service.UserService) *userController {
	return &userController{
		service: service,
	}
}

var _ api.APIGroup = (*userController)(nil)

func (c *userController) ListAPIs() []*api.API {
	return []*api.API{
		api.NewSimpleAPI("POST /api/users", c.createUser),
		api.NewSimpleAPI("GET /api/users/{userID}", c.getUser),
		api.NewSimpleAPI("PUT /api/users/{userID}", c.updateUser),
		api.NewSimpleAPI("PUT /api/users/{userID}/tokens", c.updateUserTokens),
		api.NewSimpleAPI("DELETE /api/users/{userID}", c.deleteUser),
	}
}

func (c *userController) createUser(w http.ResponseWriter, r *http.Request) error {
	session := r.Context().Value(api.SessionKey{}).(*api.Session)

	param := &domain.User{}
	if err := json.NewDecoder(r.Body).Decode(param); err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}
	defer r.Body.Close()
	param.AccessToken = session.Token.AccessToken
	param.RefreshToken = session.Token.RefreshToken
	param.NotionUserID = session.NotionUserID

	id, err := c.service.CreateUser(r.Context(), param)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	session.UserID = id
	api.SessionStore.Set(session.NotionUserID.String(), session)
	http.SetCookie(w, &http.Cookie{
		Name:     "userID",
		Value:    id.String(),
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
	})

	return api.ResponseStatusCode(w, http.StatusCreated, "success to create user")
}

func (c *userController) getUser(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	session := r.Context().Value(api.SessionKey{}).(*api.Session)
	if session.UserID != userUID {
		return api.ErrInvalidSession
	}

	user, err := c.service.GetUser(r.Context(), userUID)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	return api.ResponseJSON(w, user)
}

func (c *userController) updateUser(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	session := r.Context().Value(api.SessionKey{}).(*api.Session)
	if session.UserID != userUID {
		return api.ErrInvalidSession
	}

	param := &struct {
		Nickname string `json:"nickname"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(param); err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}
	defer r.Body.Close()

	if err := c.service.UpdateNickname(r.Context(), userUID, param.Nickname); err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	return api.ResponseStatusCode(w, http.StatusOK, "success to update user")
}

func (c *userController) updateUserTokens(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	session := r.Context().Value(api.SessionKey{}).(*api.Session)
	if session.UserID != userUID {
		return api.ErrInvalidSession
	}

	param := &struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}{}
	if err := json.NewDecoder(r.Body).Decode(param); err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}
	defer r.Body.Close()

	if err := c.service.UpdateTokens(r.Context(), userUID, param.AccessToken, param.RefreshToken); err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	return api.ResponseStatusCode(w, http.StatusOK, "success to update tokens")
}

func (c *userController) deleteUser(w http.ResponseWriter, r *http.Request) error {
	userID := r.PathValue("userID")

	userUID, err := uuid.Parse(userID)
	if err != nil {
		return api.NewError(http.StatusBadRequest, api.WithError(err))
	}

	session := r.Context().Value(api.SessionKey{}).(*api.Session)
	if session.UserID != userUID {
		return api.ErrInvalidSession
	}

	if err := c.service.DeleteUser(r.Context(), userUID); err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	return api.ResponseStatusCode(w, http.StatusOK, "success to delete user")
}
