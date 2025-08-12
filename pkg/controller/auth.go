package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/api"
	"github.com/the-end-of-the-human-era-has-arrived/2025-oss-dev-competition-backend/pkg/config"
)

// Notion OAuth token response (owner=user 기준)
type notionTokenResp struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	BotID         string `json:"bot_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	Owner         struct {
		Type string       `json:"type"`
		User *notionUser  `json:"user,omitempty"`
	} `json:"owner"`
}

type notionUser struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
}

type AuthController struct {
	oauth *config.OAuthConfig
}

func NewAuthController(_ any, oauth *config.OAuthConfig) (api.APIGroup, error) {
	// userSvc 안 써도 시그니처 호환을 위해 받도록 유지
	return &AuthController{oauth: oauth}, nil
}

func (a *AuthController) ListAPIs() []*api.API {
	return []*api.API{
		api.NewSimpleAPI("GET /login", a.login),
		api.NewSimpleAPI("GET /auth/notion/callback", a.callback),
		api.NewSimpleAPI("GET /api/me", a.me),
		api.NewSimpleAPI("POST /api/logout", a.logout),
	}
}

func (a *AuthController) login(w http.ResponseWriter, r *http.Request) error {
	authURL := buildAuthorizeURL(a.oauth)
	//state는 요청마다 난수로 만드는 게 이상적이지만, 세션스토리지가 없으므로 쿠키에 저장
	state := a.oauth.State
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   10 * 60, // 10min
	})
	http.Redirect(w, r, authURL, http.StatusFound)
	return nil
}

func (a *AuthController) callback(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		return api.NewError(http.StatusBadRequest, api.WithMessage("missing code"))
	}
	stateCookie, _ := r.Cookie("oauth_state")
	if stateCookie == nil || stateCookie.Value != a.oauth.State || state != a.oauth.State {
		return api.NewError(http.StatusBadRequest, api.WithMessage("invalid state"))
	}

	// 토큰 교환
	tok, err := exchangeToken(a.oauth, code)
	if err != nil {
		return err
	}

	// userId 결정: token.owner.user.id 우선 사용
	userID := ""
	if tok.Owner.User != nil && tok.Owner.User.ID != "" {
		userID = tok.Owner.User.ID
	}
	// 그래도 비어있으면 users/me 또는 users/{id} 로 보강이 가능하지만
	// owner=user 흐름이면 위 값이 오는 게 일반적.

	// 쿠키 저장 (로컬 개발: SameSite=None 필요 → cross-site XHR에 쿠키 포함)
	setCookie := func(name, val string, httpOnly bool) {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    url.QueryEscape(val),
			Path:     "/",
			HttpOnly: httpOnly,
			// 개발 편의: localhost에서는 Secure 미설정 허용. 배포 시 true 권장.
			Secure:   false,
			SameSite: http.SameSiteNoneMode,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
		})
	}

	setCookie("userId", userID, true)
	setCookie("notionAccessToken", tok.AccessToken, true)
	setCookie("workspaceName", tok.WorkspaceName, false)

	// 프론트로 리다이렉트
	fe := a.oauth.FrontendOrigin
	if fe == "" {
		fe = "http://localhost:3000"
	}
	http.Redirect(w, r, fe, http.StatusFound)
	return nil
}

func (a *AuthController) me(w http.ResponseWriter, r *http.Request) error {
	// /api 경로이므로 mux에서 userId 존재 검증 완료, 컨텍스트에서 꺼내도 되고 쿠키에서 읽어도 OK
	userIDCookie, _ := r.Cookie("userId")
	accessCookie, _ := r.Cookie("notionAccessToken")
	wsCookie, _ := r.Cookie("workspaceName")

	if userIDCookie == nil || accessCookie == nil {
		return api.ResponseJSON(w, map[string]any{
			"isLoggedIn": false,
		})
	}
	userID, _ := url.QueryUnescape(userIDCookie.Value)
	access, _ := url.QueryUnescape(accessCookie.Value)
	workspace, _ := url.QueryUnescape(wsCookie.Value)

	// 유저 이름 보강: Notion /v1/users/{id}
	userName := ""
	if userID != "" && access != "" {
		if name, err := fetchNotionUserName(access, userID); err == nil {
			userName = name
		}
	}

	return api.ResponseJSON(w, map[string]any{
		"isLoggedIn":    true,
		"userId":        userID,
		"userName":      userName,
		"workspaceName": workspace,
	})
}

func (a *AuthController) logout(w http.ResponseWriter, r *http.Request) error {
	clear := func(name string) {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			SameSite: http.SameSiteNoneMode,
			Secure:   false,
		})
	}
	for _, n := range []string{"userId", "notionAccessToken", "workspaceName", "oauth_state"} {
		clear(n)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// -------- helpers --------

func buildAuthorizeURL(oc *config.OAuthConfig) string {
	q := url.Values{}
	q.Set("client_id", oc.ClientID)
	q.Set("response_type", "code")
	q.Set("owner", "user")
	q.Set("redirect_uri", oc.RedirectURI)
	q.Set("state", oc.State)

	u := url.URL{
		Scheme:   "https",
		Host:     "api.notion.com",
		Path:     "/v1/oauth/authorize",
		RawQuery: q.Encode(),
	}
	return u.String()
}

func exchangeToken(oc *config.OAuthConfig, code string) (*notionTokenResp, error) {
	body := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": oc.RedirectURI,
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "https://api.notion.com/v1/oauth/token", bytes.NewReader(b))
	cred := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", oc.ClientID, oc.ClientSecret)))
	req.Header.Set("Authorization", "Basic "+cred)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, api.NewError(http.StatusBadGateway, api.WithError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var e map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return nil, api.NewError(resp.StatusCode, api.WithMessage(fmt.Sprintf("token exchange failed: %v", e)))
	}

	var tok notionTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, api.NewError(http.StatusBadGateway, api.WithError(err))
	}
	return &tok, nil
}

func fetchNotionUserName(accessToken, userID string) (string, error) {
	req, _ := http.NewRequest("GET", "https://api.notion.com/v1/users/"+url.PathEscape(userID), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	// 안정적인 버전 헤더(필요 시 최신으로 조정)
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("notion users/%s failed: %s", userID, resp.Status)
	}
	var u notionUser
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}
	return u.Name, nil
}
