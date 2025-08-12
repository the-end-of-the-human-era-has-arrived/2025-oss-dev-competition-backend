package domain

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id,omitempty"`
	Nickname     string    `json:"nickname"`
	NotionUserID uuid.UUID `json:"notion_user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}
