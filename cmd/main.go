// backend/cmd/main.go
package main

import (
    "bufio"
    "bytes"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "strings"
)

// TokenResponse represents the OAuth token exchange response.
type TokenResponse struct {
    AccessToken   string `json:"access_token"`
    RefreshToken  string `json:"refresh_token"`
    BotID         string `json:"bot_id"`
    WorkspaceName string `json:"workspace_name"`
}

// in-memory session store: sessionID → TokenResponse
var sessionStore = map[string]TokenResponse{}

// loadEnv reads a .env file at path and sets os.Getenv for each KEY=VALUE line.
func loadEnv(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }
        key := strings.TrimSpace(parts[0])
        val := strings.TrimSpace(parts[1])
        if strings.HasPrefix(val, `"`) && strings.HasSuffix(val, `"`) {
            val = strings.Trim(val, `"`)
        }
        os.Setenv(key, val)
    }
    return scanner.Err()
}

// generateSessionID returns a new random hex string.
func generateSessionID() string {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil {
        return ""
    }
    return fmt.Sprintf("%x", b)
}

func main() {
    // 1) load .env from backend/.env
    _ = loadEnv("../.env")
    _ = loadEnv(".env")

    // 2) read env vars
    clientID := os.Getenv("OAUTH_CLIENT_ID")
    clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
    redirectURI := os.Getenv("NOTION_REDIRECT_URI")
    if clientID == "" || clientSecret == "" || redirectURI == "" {
        log.Fatal("환경변수 OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, NOTION_REDIRECT_URI를 설정하세요")
    }

    // 3) /auth/notion: redirect to Notion OAuth consent
    http.HandleFunc("/auth/notion", func(w http.ResponseWriter, r *http.Request) {
        state := generateSessionID() // or any CSRF token
        authURL, err := url.Parse("https://api.notion.com/v1/oauth/authorize")
        if err != nil {
            http.Error(w, "authorize URL 생성 실패", http.StatusInternalServerError)
            return
        }
        q := authURL.Query()
        q.Set("owner", "user")
        q.Set("client_id", clientID)
        q.Set("redirect_uri", redirectURI)
        q.Set("response_type", "code")
        q.Set("state", state)
        authURL.RawQuery = q.Encode()

        http.Redirect(w, r, authURL.String(), http.StatusTemporaryRedirect)
    })

    // 4) /auth/notion/callback: exchange code for token, set session, redirect to /list
    http.HandleFunc("/auth/notion/callback", func(w http.ResponseWriter, r *http.Request) {
        code := r.URL.Query().Get("code")
        if code == "" {
            http.Error(w, "authorization code 누락", http.StatusBadRequest)
            return
        }
        // TODO: state 검증

        tokenURL := "https://api.notion.com/v1/oauth/token"
        payload := map[string]string{
            "grant_type":   "authorization_code",
            "code":         code,
            "redirect_uri": redirectURI,
        }
        body, _ := json.Marshal(payload)
        req, err := http.NewRequest("POST", tokenURL, bytes.NewReader(body))
        if err != nil {
            http.Error(w, "토큰 요청 생성 실패", http.StatusInternalServerError)
            return
        }
        basicAuth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
        req.Header.Set("Authorization", "Basic "+basicAuth)
        req.Header.Set("Content-Type", "application/json")

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            http.Error(w, "토큰 요청 실패", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            data, _ := io.ReadAll(resp.Body)
            http.Error(w, "토큰 교환 오류: "+string(data), resp.StatusCode)
            return
        }

        var tr TokenResponse
        if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
            http.Error(w, "응답 디코드 실패", http.StatusInternalServerError)
            return
        }

        // 세션 생성 및 저장
        sessionID := generateSessionID()
        sessionStore[sessionID] = tr
        http.SetCookie(w, &http.Cookie{
            Name:     "session_id",
            Value:    sessionID,
            Path:     "/",
            HttpOnly: true,
        })

        // /list로 리다이렉트
        http.Redirect(w, r, "/list", http.StatusSeeOther)
    })

    // 5) /list: 세션에서 토큰 꺼내 Notion Search API 호출하여 페이지 목록 반환
    http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_id")
        if err != nil {
            http.Error(w, "세션이 없습니다. 다시 로그인해주세요.", http.StatusUnauthorized)
            return
        }
        tr, ok := sessionStore[cookie.Value]
        if !ok {
            http.Error(w, "유효하지 않은 세션입니다.", http.StatusUnauthorized)
            return
        }

        // Notion Search API 호출
        searchURL := "https://api.notion.com/v1/search"
        reqBody := map[string]interface{}{
            "filter": map[string]string{
                "property": "object",
                "value":    "page",
            },
        }
        b, _ := json.Marshal(reqBody)
        req, _ := http.NewRequest("POST", searchURL, bytes.NewReader(b))
        req.Header.Set("Authorization", "Bearer "+tr.AccessToken)
        req.Header.Set("Notion-Version", "2022-06-28")
        req.Header.Set("Content-Type", "application/json")

        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            http.Error(w, "Notion API 호출 실패", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            data, _ := io.ReadAll(resp.Body)
            http.Error(w, "Notion 검색 오류: "+string(data), resp.StatusCode)
            return
        }

        // 결과를 그대로 클라이언트에 반환
        w.Header().Set("Content-Type", "application/json")
        io.Copy(w, resp.Body)
    })

    log.Println("서버 구동 ➡ http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
