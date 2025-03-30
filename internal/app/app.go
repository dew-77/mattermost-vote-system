package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/sirupsen/logrus"
	"github.com/dew-77/mattermost-vote-system/internal/config"
	"github.com/dew-77/mattermost-vote-system/internal/mattermost"
	"github.com/dew-77/mattermost-vote-system/internal/repository"
)

type App struct {
	config     *config.Config
	logger     *logrus.Logger
	mmClient   *mattermost.Client
	repository repository.PollRepository
}

func NewApp(cfg *config.Config, logger *logrus.Logger, mmClient *mattermost.Client, repo repository.PollRepository) *App {
	return &App{
		config:     cfg,
		logger:     logger,
		mmClient:   mmClient,
		repository: repo,
	}
}

func (a *App) Start() error {
	wsClient, err := a.mmClient.GetWebSocketClient()
	if err != nil {
		return err
	}
	
	wsClient.Listen()
	
	a.logger.Info("Bot started and listening for events")
	
	for {
		select {
		case event := <-wsClient.EventChannel:
			a.handleWebSocketEvent(event)
		}
	}
}

func (a *App) handleWebSocketEvent(event *model.WebSocketEvent) {
	if event.EventType() != model.WebsocketEventPosted {
		return
	}
	
	if event.GetBroadcast().UserId == a.mmClient.GetBotUserID() {
		return
	}
	
	postData, ok := event.GetData()["post"].(string)
	if !ok {
		return
	}
	
	var post model.Post
	if err := json.Unmarshal([]byte(postData), &post); err != nil {
		return
	}
	
	channelType, ok := event.GetData()["channel_type"].(string)
	if !ok {
		channelType = ""
	}
	
	if !strings.Contains(post.Message, fmt.Sprintf("@%s", a.mmClient.GetBotUserID())) && 
	   channelType != string(model.ChannelTypeDirect) {
		return
	}
	
	a.logger.WithFields(logrus.Fields{
		"user_id":    event.GetBroadcast().UserId,
		"channel_id": event.GetBroadcast().ChannelId,
		"message":    post.Message,
	}).Info("Received message")
	
	a.handleCommand(event.GetBroadcast().UserId, event.GetBroadcast().ChannelId, post.Message)
}

func (a *App) handleCommand(userID, channelID, message string) {
	message = strings.ReplaceAll(message, fmt.Sprintf("@%s", a.mmClient.GetBotUserID()), "")
	message = strings.TrimSpace(message)
	
	parts := strings.Fields(message)
	if len(parts) == 0 {
		a.replyHelp(channelID)
		return
	}
	
	command := strings.ToLower(parts[0])
	
	switch command {
	case "create", "new", "poll":
		a.handleCreatePoll(userID, channelID, parts[1:])
	case "vote":
		a.handleVote(userID, channelID, parts[1:])
	case "results":
		a.handleResults(channelID, parts[1:])
	case "finish":
		a.handleFinishPoll(userID, channelID, parts[1:])
	case "delete":
		a.handleDeletePoll(userID, channelID, parts[1:])
	case "help":
		a.replyHelp(channelID)
	default:
		a.replyHelp(channelID)
	}
}

func (a *App) replyHelp(channelID string) {
	helpText := `### Команды голосования:
- **create "Заголовок" "Вариант 1" "Вариант 2" ...** - Создать новое голосование
- **vote [ID голосования] [номер варианта]** - Проголосовать за вариант
- **results [ID голосования]** - Показать результаты голосования
- **finish [ID голосования]** - Завершить голосование (только для создателя)
- **delete [ID голосования]** - Удалить голосование (только для создателя)
- **help** - Показать эту справку`

	a.mmClient.CreatePost(channelID, helpText)
}
