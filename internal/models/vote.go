package models

import (
	"time"
)

type Vote struct {
	PollID    string    `json:"poll_id"`
	UserID    string    `json:"user_id"`
	OptionIdx int       `json:"option_idx"`
	VotedAt   time.Time `json:"voted_at"`
}