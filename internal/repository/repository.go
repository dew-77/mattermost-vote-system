package repository

import (
	"github.com/dew-77/mattermost-vote-system/internal/models"
)

type PollRepository interface {
	CreatePoll(poll models.Poll) error
	GetPoll(pollID string) (models.Poll, error)
	UpdatePoll(poll models.Poll) error
	DeletePoll(pollID string) error

	AddVote(vote models.Vote) error
	GetVotes(pollID string) ([]models.Vote, error)

	GetPollResults(pollID string) (models.PollResults, error)
}
