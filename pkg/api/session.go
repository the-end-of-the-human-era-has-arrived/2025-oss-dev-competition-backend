package api

import (
	"sync"

	"github.com/google/uuid"
)

// NotionUserID를 key로 session 저장
var SessionStore = &inMemoryStore{
	store: make(map[string]*Session, 0),
}

// ***********************************
// TEST
// ***********************************
// var SessionStore = &inMemoryStore{
// 	store: map[string]*Session{
// 		"3aeac03f-0645-4d10-a334-a5905c031c12": {
// 			UserID:       uuid.MustParse("3aeac03f-0645-4d10-a334-a5905c031c12"),
// 			NotionUserID: uuid.MustParse("97811864-e95f-41ee-8faf-e5dfce7d0326"),
// 			Token: &Token{
// 				AccessToken:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lI",
// 				RefreshToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lI",
// 				BotID:         "3399b2c7-a7ad-4778-a803-08262e4ed808",
// 				WorkspaceName: "test",
// 			},
// 		},
// 	},
// }

type inMemoryStore struct {
	mu    sync.RWMutex
	store map[string]*Session
}

func (s *inMemoryStore) Get(key string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.store[key]
	if !ok {
		return nil, false
	}
	copied := *session
	return &copied, ok
}

func (s *inMemoryStore) Set(key string, session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = session
}

type Session struct {
	UserID       uuid.UUID `json:"user_id"`
	NotionUserID uuid.UUID `json:"notion_user_id"`
	Token        *Token    `json:"token"`
}

type Token struct {
	AccessToken   string `json:"access_token"`
	RefreshToken  string `json:"refresh_token"`
	BotID         string `json:"bot_id"`
	WorkspaceName string `json:"workspace_name"`
}
