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
// 			NotionUserID: uuid.MustParse("3aeac03f-0645-4d10-a334-a5905c031c12"),
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
	UserID       uuid.UUID `json:"userID"`
	NotionUserID uuid.UUID `json:"notionUserID"`
	Token        *Token    `json:"token"`
}

type Token struct {
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken"`
	BotID         string `json:"botID"`
	WorkspaceName string `json:"workspaceName"`
}

// 컨트롤러/미들웨어에서 컨텍스트로 세션을 주고받는 키
type SessionKey struct{}