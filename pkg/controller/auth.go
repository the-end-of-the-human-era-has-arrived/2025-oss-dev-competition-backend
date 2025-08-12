package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/config"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/domain"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/service"
)

const (
	tokenURL = "https://api.notion.com/v1/oauth/token"
)

type authController struct {
	service      *service.UserService
	clientID     string
	clientSecret string
	authURL      string
	state        string
}

func NewAuthController(
	service *service.UserService,
	cfg *config.OAuthConfig,
) (*authController, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, errors.New("required oauth config")
	}

	return &authController{
		service:      service,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		authURL:      cfg.AuthURL,
		state:        cfg.State,
	}, nil
}

var _ api.APIGroup = (*authController)(nil)

func (c *authController) ListAPIs() []*api.API {
	return []*api.API{
		api.NewSimpleAPI("GET /auth/notion", c.processNotionAuth),
		api.NewSimpleAPI("GET /auth/notion/callback", c.processNotionAuthCallback),
		api.NewSimpleAPI("GET /api/session/status", c.getSessionStatus),
	}
}

func (c *authController) processNotionAuth(w http.ResponseWriter, r *http.Request) error {
	aURL, err := url.Parse(c.authURL)
	if err != nil {
		return api.NewError(
			http.StatusInternalServerError,
			api.WithMessage("fail to parse auth url"),
			api.WithError(err),
		)
	}

	q := aURL.Query()
	q.Set("state", c.state)
	aURL.RawQuery = q.Encode()

	http.Redirect(w, r, aURL.String(), http.StatusTemporaryRedirect)
	return nil
}

func (c *authController) processNotionAuthCallback(w http.ResponseWriter, r *http.Request) error {
	queryState := r.URL.Query().Get("state")
	if queryState != c.state {
		return api.NewError(http.StatusBadRequest, api.WithMessage("invalid state"))
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return api.NewError(http.StatusBadRequest, api.WithMessage("authorization code 누락"))
	}

	tok, err := c.requestToken(
		code,
		fmt.Sprintf("http://%s/auth/notion/callback", r.Host),
	)
	if err != nil {
		return api.NewError(
			http.StatusInternalServerError,
			api.WithMessage("fail to request token"),
			api.WithError(err),
		)
	}

	user, err := c.getNotionUser(tok.AccessToken)
	if err != nil {
		return api.NewError(
			http.StatusInternalServerError,
			api.WithMessage("fail to get notion user"),
			api.WithError(err),
		)
	}
	notionUserID := uuid.MustParse(user.ID)

	requestID, err := uuid.NewRandom()
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}
	ctx := context.WithValue(r.Context(), api.RequestIDKey{}, requestID)

	id, err := c.service.CreateUser(ctx, &domain.User{
		Nickname:     user.Name,
		NotionUserID: notionUserID,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
	})
	if err != nil {
		return api.NewError(http.StatusInternalServerError, api.WithError(err))
	}

	// 세션 생성 및 저장
	session := &api.Session{
		UserID:       id,
		NotionUserID: notionUserID,
		Token:        tok,
	}

	api.SessionStore.Set(id.String(), session)
	http.SetCookie(w, &http.Cookie{
		Name:     "sessionID",
		Value:    id.String(),
		Path:     "/",
		HttpOnly: true,
	})

	return api.ResponseStatusCode(w, http.StatusOK, "success to login")
}

func (c *authController) requestToken(code, redirectURI string) (*api.Token, error) {
	payload := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": redirectURI,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", tokenURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	basicAuth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+basicAuth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return nil, errors.New(string(data))
	}

	tok := &api.Token{}
	if err := json.NewDecoder(resp.Body).Decode(tok); err != nil {
		return nil, err
	}

	return tok, nil
}

type notionUser struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Person struct {
		Email string `json:"email"`
	} `json:"person"`
}

type notionUserResp struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
	Bot       struct {
		Owner struct {
			Type string      `json:"type"`
			User *notionUser `json:"user"`
		} `json:"owner"`
	} `json:"bot"`
}

func (c *authController) getNotionUser(accessToken string) (*notionUser, error) {
	req, err := http.NewRequest("GET", "https://api.notion.com/v1/users/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	data := &notionUserResp{}
	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return data.Bot.Owner.User, nil
}

func (c *authController) getSessionStatus(w http.ResponseWriter, r *http.Request) error {
	session := r.Context().Value(api.SessionKey{}).(*api.Session)

	response := map[string]interface{}{
		"authenticated":  true,
		"user_id":        session.UserID,
		"notion_user_id": session.NotionUserID,
	}

	return api.ResponseJSON(w, response)
}
