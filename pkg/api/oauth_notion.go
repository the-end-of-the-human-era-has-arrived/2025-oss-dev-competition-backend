package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
)

type notionTokenResp struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	BotID         string `json:"bot_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	Owner         struct {
		Type string      `json:"type"`
		User *notionUser `json:"user,omitempty"`
	} `json:"owner"`
}

type notionUser struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
}

type NotionAuthGroup struct {
	clientID       string
	clientSecret   string
	redirectURI    string
	state          string
	frontendOrigin string
}

func NewNotionAuthGroupFromEnv() *NotionAuthGroup {
	return &NotionAuthGroup{
		clientID:       os.Getenv("OAUTH_CLIENT_ID"),
		clientSecret:   os.Getenv("OAUTH_CLIENT_SECRET"),
		redirectURI:    getenvDefault("OAUTH_REDIRECT_URI", "http://localhost:8080/auth/notion/callback"),
		state:          getenvDefault("STATE", "dev-state"),
		frontendOrigin: getenvDefault("FRONTEND_ORIGIN", "http://localhost:3000"),
	}
}

func (g *NotionAuthGroup) ListAPIs() []*API {
	return []*API{
		NewSimpleAPI("GET /login", g.login),
		NewSimpleAPI("GET /auth/notion/callback", g.callback),
		NewSimpleAPI("GET /api/me", g.me),
		NewSimpleAPI("POST /api/logout", g.logout),
	}
}

func (g *NotionAuthGroup) login(w http.ResponseWriter, r *http.Request) error {
	q := url.Values{}
	q.Set("client_id", g.clientID)
	q.Set("response_type", "code")
	q.Set("owner", "user")
	q.Set("redirect_uri", g.redirectURI)
	q.Set("state", g.state)

	u := url.URL{
		Scheme:   "https",
		Host:     "api.notion.com",
		Path:     "/v1/oauth/authorize",
		RawQuery: q.Encode(),
	}
	http.Redirect(w, r, u.String(), http.StatusFound)
	return nil
}

func (g *NotionAuthGroup) callback(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		return NewError(http.StatusBadRequest, WithMessage("missing code"))
	}
	if state != g.state {
		return NewError(http.StatusBadRequest, WithMessage("invalid state"))
	}

	// 토큰 교환
	tok, err := g.exchangeToken(code)
	if err != nil {
		return err
	}

	// Notion 유저 ID 확인
	if tok.Owner.User == nil || tok.Owner.User.ID == "" {
		return NewError(http.StatusBadGateway, WithMessage("no notion user id"))
	}
	notionUIDStr := tok.Owner.User.ID
	notionUID, err := uuid.Parse(notionUIDStr)
	if err != nil {
		return NewError(http.StatusBadGateway, WithError(fmt.Errorf("invalid notion user id: %w", err)))
	}

	// 세션 생성/저장 (키 = NotionUserID 문자열)
	sess := &Session{
		UserID:       uuid.New(), // 우리 서비스 사용자 ID(필요 시 교체)
		NotionUserID: notionUID,
		Token: &Token{
			AccessToken:   tok.AccessToken,
			RefreshToken:  tok.RefreshToken,
			BotID:         tok.BotID,
			WorkspaceName: tok.WorkspaceName,
		},
	}
	SessionStore.Set(notionUIDStr, sess)

	// 쿠키 세팅: notionUserId (3000↔8080 교차요청 고려)
	http.SetCookie(w, &http.Cookie{
		Name:     NotionUserIDCookie,
		Value:    url.QueryEscape(notionUIDStr),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode, // 크로스 사이트 fetch 시 필요
		Secure:   false,                 // 배포 시 true 권장(HTTPS)
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	// 프론트로 이동
	http.Redirect(w, r, g.frontendOrigin, http.StatusFound)
	return nil
}

func (g *NotionAuthGroup) me(w http.ResponseWriter, r *http.Request) error {
	// /api/* 경로는 mux에서 이미 세션을 컨텍스트에 심어줌
	v := r.Context().Value(SessionKey{})
	if v == nil {
		return ResponseJSON(w, map[string]any{"isLoggedIn": false})
	}
	sess := v.(*Session)
	return ResponseJSON(w, map[string]any{
		"isLoggedIn":    true,
		"userId":        sess.UserID.String(),
		"notionUserId":  sess.NotionUserID.String(),
		"workspaceName": sess.Token.WorkspaceName,
	})
}

func (g *NotionAuthGroup) logout(w http.ResponseWriter, r *http.Request) error {
	// 쿠키에서 notionUserId 찾아 세션 무효화
	if c, err := r.Cookie(NotionUserIDCookie); err == nil && c.Value != "" {
		if v, err2 := url.QueryUnescape(c.Value); err2 == nil {
			// 삭제 메서드가 없으므로 nil 세팅으로 무효화( mux에서 nil 방지 처리 )
			SessionStore.Set(v, nil)
		}
	}
	// 쿠키 만료
	http.SetCookie(w, &http.Cookie{
		Name:     NotionUserIDCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteNoneMode,
		Secure:   false,
	})
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// --- helpers ---

func (g *NotionAuthGroup) exchangeToken(code string) (*notionTokenResp, error) {
	body := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": g.redirectURI,
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.notion.com/v1/oauth/token", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	cred := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", g.clientID, g.clientSecret)))
	req.Header.Set("Authorization", "Basic "+cred)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, NewError(http.StatusBadGateway, WithError(err))
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var e map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return nil, NewError(resp.StatusCode, WithMessage(fmt.Sprintf("token exchange failed: %v", e)))
	}

	var tok notionTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return nil, NewError(http.StatusBadGateway, WithError(err))
	}
	return &tok, nil
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
