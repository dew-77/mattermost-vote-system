package models

import (
	"time"
)

type Poll struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Options     []string  `json:"options"`
	CreatorID   string    `json:"creator_id"`
	ChannelID   string    `json:"channel_id"`
	CreatedAt   time.Time `json:"created_at"`
	FinishedAt  time.Time `json:"finished_at,omitempty"`
	IsFinished  bool      `json:"is_finished"`
	PostID      string    `json:"post_id"`
}

type PollResults struct {
	Poll    Poll             `json:"poll"`
	Results map[string]int   `json:"results"`
	Voters  map[string]string `json:"voters"`
}