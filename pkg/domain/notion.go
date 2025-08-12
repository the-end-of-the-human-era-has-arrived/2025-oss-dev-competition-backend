package domain

import "github.com/google/uuid"

type NotionPage struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Content      string    `json:"content"`
	NotionURL    string    `json:"notion_url"`
	NotionPageID uuid.UUID `json:"notion_page_id"`
	Summary      string    `json:"summary"`
}
