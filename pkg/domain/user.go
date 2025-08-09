package domain

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id,omitempty"`
	Nickname     string    `json:"nickname"`
	NotionUserID uuid.UUID `json:"notionUserID"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
}
