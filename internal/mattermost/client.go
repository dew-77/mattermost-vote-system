package mattermost

import (
	"fmt"
	"strings"

	"github.com/dew-77/mattermost-vote-system/internal/config"
	"github.com/mattermost/mattermost-server/v6/model"
)

type Client struct {
	client    *model.Client4
	botUserID string
	teamID    string
	config    *config.MattermostConfig
}

func NewClient(cfg *config.MattermostConfig) (*Client, error) {
	fmt.Printf("Connecting to Mattermost at: %s\n", cfg.ServerURL)
	fmt.Printf("Team name: %s\n", cfg.TeamName)
	fmt.Printf("Bot User ID: %s\n", cfg.BotUserID)

	client := model.NewAPIv4Client(cfg.ServerURL)
	client.SetToken(cfg.Token)

	_, resp, err := client.GetMe("")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Mattermost: %v", err)
	}
	if resp != nil && resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to connect to Mattermost: status code %d", resp.StatusCode)
	}

	team, resp, err := client.GetTeamByName(cfg.TeamName, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %v", err)
	}
	if resp != nil && resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get team: status code %d", resp.StatusCode)
	}

	return &Client{
		client:    client,
		botUserID: cfg.BotUserID,
		teamID:    team.Id,
		config:    cfg,
	}, nil
}

func (c *Client) CreatePost(channelID, message string) (*model.Post, error) {
	post := &model.Post{
		ChannelId: channelID,
		Message:   message,
	}

	post, resp, err := c.client.CreatePost(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %v", err)
	}
	if resp != nil && resp.StatusCode != 201 {
		return nil, fmt.Errorf("failed to create post: status code %d", resp.StatusCode)
	}

	return post, nil
}

func (c *Client) UpdatePost(post *model.Post) (*model.Post, error) {
	post, resp, err := c.client.UpdatePost(post.Id, post)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %v", err)
	}
	if resp != nil && resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to update post: status code %d", resp.StatusCode)
	}

	return post, nil
}

func (c *Client) DeletePost(postID string) error {
	resp, err := c.client.DeletePost(postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %v", err)
	}
	if resp != nil && resp.StatusCode != 200 {
		return fmt.Errorf("failed to delete post: status code %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetUser(userID string) (*model.User, error) {
	user, resp, err := c.client.GetUser(userID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if resp != nil && resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get user: status code %d", resp.StatusCode)
	}

	return user, nil
}

func (c *Client) GetBotUserID() string {
	return c.botUserID
}

func (c *Client) GetTeamID() string {
	return c.teamID
}

func (c *Client) GetChannelByName(channelName string) (*model.Channel, error) {
	channel, resp, err := c.client.GetChannelByName(channelName, c.teamID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %v", err)
	}
	if resp != nil && resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get channel: status code %d", resp.StatusCode)
	}

	return channel, nil
}

func (c *Client) GetWebSocketClient() (*model.WebSocketClient, error) {
	wsURL := c.config.ServerURL
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	if !strings.HasSuffix(wsURL, "/api/v4/websocket") {
		if !strings.HasSuffix(wsURL, "/") {
			wsURL += "/"
		}
		wsURL += "api/v4/websocket"
	}

	fmt.Printf("Connecting to WebSocket at: %s\n", wsURL)

	wsClient, err := model.NewWebSocketClient(wsURL, c.client.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebSocket client: %v", err)
	}

	go wsClient.Listen()

	return wsClient, nil
}
