package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/tarantool/go-tarantool"
	"github.com/dew-77/mattermost-vote-system/internal/config"
	"github.com/dew-77/mattermost-vote-system/internal/models"
)

type TarantoolRepository struct {
	conn   *tarantool.Connection
	config *config.TarantoolConfig
}

func NewTarantoolRepository(cfg *config.TarantoolConfig) (*TarantoolRepository, error) {
	log.Printf("Attempting to connect to Tarantool at %s:%d with user '%s'", cfg.Host, cfg.Port, cfg.User)
	
	opts := tarantool.Opts{
		User:      cfg.User,
		Pass:      cfg.Password,
		Timeout:   5 * time.Second,
		Reconnect: 1 * time.Second,
	}
	
	log.Printf("Connection options: Timeout=%v, Reconnect=%v", opts.Timeout, opts.Reconnect)
	
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Printf("Connection string: %s", addr)
	
	log.Printf("Calling tarantool.Connect()...")
	conn, err := tarantool.Connect(addr, opts)
	
	if err != nil {
		log.Printf("ERROR: Failed to connect to Tarantool: %v", err)
		return nil, fmt.Errorf("failed to connect to Tarantool: %w", err)
	}
	
	log.Printf("Successfully connected to Tarantool")
	
	log.Printf("Testing connection with Ping...")
	resp, err := conn.Ping()
	if err != nil {
		log.Printf("ERROR: Ping failed: %v", err)
		conn.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}
	log.Printf("Ping successful: %v", resp)
	
	log.Printf("Checking for required spaces...")
	spaces, err := conn.Select("_space", "name", 0, 1, tarantool.IterEq, []interface{}{"polls"})
	if err != nil {
		log.Printf("ERROR: Failed to check for 'polls' space: %v", err)
	} else if len(spaces.Data) == 0 {
		log.Printf("WARNING: 'polls' space not found")
	} else {
		log.Printf("'polls' space found")
	}
	
	spaces, err = conn.Select("_space", "name", 0, 1, tarantool.IterEq, []interface{}{"votes"})
	if err != nil {
		log.Printf("ERROR: Failed to check for 'votes' space: %v", err)
	} else if len(spaces.Data) == 0 {
		log.Printf("WARNING: 'votes' space not found")
	} else {
		log.Printf("'votes' space found")
	}
	
	return &TarantoolRepository{
		conn:   conn,
		config: cfg,
	}, nil
}

func (r *TarantoolRepository) CreatePoll(poll models.Poll) error {
	log.Printf("Creating poll with ID: %s", poll.ID)
	resp, err := r.conn.Insert("polls", []interface{}{
		poll.ID,
		poll.Title,
		poll.Options,
		poll.CreatorID,
		poll.ChannelID,
		poll.CreatedAt,
		poll.FinishedAt,
		poll.IsFinished,
		poll.PostID,
	})
	
	if err != nil {
		log.Printf("ERROR: Failed to create poll: %v", err)
		return fmt.Errorf("failed to create poll: %w", err)
	}
	
	log.Printf("Poll created successfully: %v", resp)
	return nil
}

func (r *TarantoolRepository) GetPoll(pollID string) (models.Poll, error) {
	log.Printf("Getting poll with ID: %s", pollID)
	
	resp, err := r.conn.Select("polls", "primary", 0, 1, tarantool.IterEq, []interface{}{pollID})
	if err != nil {
		log.Printf("ERROR: Failed to get poll: %v", err)
		return models.Poll{}, fmt.Errorf("failed to get poll: %w", err)
	}
	
	if len(resp.Data) == 0 {
		log.Printf("Poll not found with ID: %s", pollID)
		return models.Poll{}, fmt.Errorf("poll not found")
	}
	
	tuples := resp.Tuples()
	if len(tuples) == 0 {
		log.Printf("Poll tuples empty for ID: %s", pollID)
		return models.Poll{}, fmt.Errorf("poll not found")
	}
	
	tuple := tuples[0]
	log.Printf("Retrieved poll tuple: %v", tuple)
	
	var poll models.Poll
	poll.ID = tuple[0].(string)
	poll.Title = tuple[1].(string)
	
	optionsInterface := tuple[2].([]interface{})
	poll.Options = make([]string, len(optionsInterface))
	for i, opt := range optionsInterface {
		poll.Options[i] = opt.(string)
	}
	
	poll.CreatorID = tuple[3].(string)
	poll.ChannelID = tuple[4].(string)
	poll.CreatedAt = tuple[5].(time.Time)
	if tuple[6] != nil {
		poll.FinishedAt = tuple[6].(time.Time)
	}
	poll.IsFinished = tuple[7].(bool)
	poll.PostID = tuple[8].(string)
	
	log.Printf("Successfully retrieved poll: %s - %s", poll.ID, poll.Title)
	return poll, nil
}

func (r *TarantoolRepository) UpdatePoll(poll models.Poll) error {
	log.Printf("Updating poll with ID: %s", poll.ID)
	
	resp, err := r.conn.Replace("polls", []interface{}{
		poll.ID,
		poll.Title,
		poll.Options,
		poll.CreatorID,
		poll.ChannelID,
		poll.CreatedAt,
		poll.FinishedAt,
		poll.IsFinished,
		poll.PostID,
	})
	
	if err != nil {
		log.Printf("ERROR: Failed to update poll: %v", err)
		return fmt.Errorf("failed to update poll: %w", err)
	}
	
	log.Printf("Poll updated successfully: %v", resp)
	return nil
}

func (r *TarantoolRepository) DeletePoll(pollID string) error {
	log.Printf("Deleting poll with ID: %s", pollID)
	
	resp, err := r.conn.Delete("polls", "primary", []interface{}{pollID})
	if err != nil {
		log.Printf("ERROR: Failed to delete poll: %v", err)
		return fmt.Errorf("failed to delete poll: %w", err)
	}
	
	log.Printf("Poll deleted successfully: %v", resp)
	return nil
}

func (r *TarantoolRepository) AddVote(vote models.Vote) error {
	log.Printf("Adding vote for poll %s by user %s for option %d", vote.PollID, vote.UserID, vote.OptionIdx)
	
	resp, err := r.conn.Delete("votes", "user_poll", []interface{}{vote.UserID, vote.PollID})
	if err != nil {
		log.Printf("ERROR: Failed to delete previous vote: %v", err)
		return fmt.Errorf("failed to delete previous vote: %w", err)
	}
	log.Printf("Previous vote deleted (if any): %v", resp)
	
	resp, err = r.conn.Insert("votes", []interface{}{
		vote.PollID,
		vote.UserID,
		vote.OptionIdx,
		vote.VotedAt,
	})
	
	if err != nil {
		log.Printf("ERROR: Failed to add vote: %v", err)
		return fmt.Errorf("failed to add vote: %w", err)
	}
	
	log.Printf("Vote added successfully: %v", resp)
	return nil
}

func (r *TarantoolRepository) GetVotes(pollID string) ([]models.Vote, error) {
	log.Printf("Getting votes for poll ID: %s", pollID)
	
	resp, err := r.conn.Select("votes", "poll", 0, 100, tarantool.IterEq, []interface{}{pollID})
	if err != nil {
		log.Printf("ERROR: Failed to get votes: %v", err)
		return nil, fmt.Errorf("failed to get votes: %w", err)
	}
	
	tuples := resp.Tuples()
	log.Printf("Retrieved %d votes for poll %s", len(tuples), pollID)
	
	votes := make([]models.Vote, len(tuples))
	
	for i, tuple := range tuples {
		votes[i] = models.Vote{
			PollID:    tuple[0].(string),
			UserID:    tuple[1].(string),
			OptionIdx: tuple[2].(int),
			VotedAt:   tuple[3].(time.Time),
		}
	}
	
	return votes, nil
}

func (r *TarantoolRepository) GetPollResults(pollID string) (models.PollResults, error) {
	log.Printf("Getting poll results for ID: %s", pollID)
	
	poll, err := r.GetPoll(pollID)
	if err != nil {
		log.Printf("ERROR: Failed to get poll for results: %v", err)
		return models.PollResults{}, fmt.Errorf("failed to get poll for results: %w", err)
	}
	
	votes, err := r.GetVotes(pollID)
	if err != nil {
		log.Printf("ERROR: Failed to get votes for results: %v", err)
		return models.PollResults{}, fmt.Errorf("failed to get votes for results: %w", err)
	}
	
	results := make(map[string]int)
	for _, option := range poll.Options {
		results[option] = 0
	}
	
	voters := make(map[string]string)
	
	for _, vote := range votes {
		if vote.OptionIdx >= 0 && vote.OptionIdx < len(poll.Options) {
			option := poll.Options[vote.OptionIdx]
			results[option]++
			voters[vote.UserID] = option
		}
	}
	
	log.Printf("Poll results calculated: %v options, %v votes", len(results), len(voters))
	return models.PollResults{
		Poll:    poll,
		Results: results,
		Voters:  voters,
	}, nil
}

func (r *TarantoolRepository) HealthCheck() error {
	log.Printf("Performing health check...")
	
	resp, err := r.conn.Ping()
	if err != nil {
		log.Printf("ERROR: Health check failed: %v", err)
		return fmt.Errorf("health check failed: %w", err)
	}
	
	log.Printf("Health check successful: %v", resp)
	return nil
}