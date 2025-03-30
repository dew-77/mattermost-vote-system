package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/dew-77/mattermost-vote-system/internal/models"
)

func (a *App) handleCreatePoll(userID, channelID string, args []string) {
	if len(args) < 3 {
		a.mmClient.CreatePost(channelID, "Ошибка: Недостаточно аргументов. Используйте: create \"Заголовок\" \"Вариант 1\" \"Вариант 2\" ...")
		return
	}
	
	fullText := strings.Join(args, " ")
	parts := splitQuoted(fullText)
	
	if len(parts) < 3 {
		a.mmClient.CreatePost(channelID, "Ошибка: Необходимо указать заголовок и минимум 2 варианта ответа в кавычках.")
		return
	}
	
	title := parts[0]
	options := parts[1:]
	
	pollID := uuid.New().String()[:8]
	
	poll := models.Poll{
		ID:         pollID,
		Title:      title,
		Options:    options,
		CreatorID:  userID,
		ChannelID:  channelID,
		CreatedAt:  time.Now(),
		IsFinished: false,
	}
	
	message := formatPollMessage(poll, nil)
	
	post, err := a.mmClient.CreatePost(channelID, message)
	if err != nil {
		a.logger.WithError(err).Error("Failed to create poll post")
		a.mmClient.CreatePost(channelID, "Ошибка при создании голосования.")
		return
	}
	
	poll.PostID = post.Id
	
	err = a.repository.CreatePoll(poll)
	if err != nil {
		a.logger.WithError(err).Error("Failed to save poll to database")
		a.mmClient.CreatePost(channelID, "Ошибка при сохранении голосования.")
		return
	}
	
	a.mmClient.CreatePost(channelID, fmt.Sprintf("Голосование создано! ID: `%s`", pollID))
}

func (a *App) handleVote(userID, channelID string, args []string) {
	if len(args) < 2 {
		a.mmClient.CreatePost(channelID, "Ошибка: Недостаточно аргументов. Используйте: vote [ID голосования] [номер варианта]")
		return
	}
	
	pollID := args[0]
	
	optionIdx, err := strconv.Atoi(args[1])
	if err != nil {
		a.mmClient.CreatePost(channelID, "Ошибка: Номер варианта должен быть числом.")
		return
	}
	
	poll, err := a.repository.GetPoll(pollID)
	if err != nil {
		a.mmClient.CreatePost(channelID, fmt.Sprintf("Ошибка: Голосование с ID `%s` не найдено.", pollID))
		return
	}
	
	if poll.IsFinished {
		a.mmClient.CreatePost(channelID, "Ошибка: Голосование уже завершено.")
		return
	}
	
	if optionIdx < 1 || optionIdx > len(poll.Options) {
		a.mmClient.CreatePost(channelID, fmt.Sprintf("Ошибка: Номер варианта должен быть от 1 до %d.", len(poll.Options)))
		return
	}
	
	vote := models.Vote{
		PollID:    pollID,
		UserID:    userID,
		OptionIdx: optionIdx - 1,
		VotedAt:   time.Now(),
	}
	
	err = a.repository.AddVote(vote)
	if err != nil {
		a.logger.WithError(err).Error("Failed to save vote")
		a.mmClient.CreatePost(channelID, "Ошибка при сохранении голоса.")
		return
	}
	
	results, err := a.repository.GetPollResults(pollID)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get poll results")
		a.mmClient.CreatePost(channelID, "Ошибка при получении результатов голосования.")
		return
	}
	
	formatPollMessage(poll, &results)
	
	user, err := a.mmClient.GetUser(userID)
	if err == nil {
		a.mmClient.CreatePost(channelID, fmt.Sprintf("@%s проголосовал за вариант %d в голосовании `%s`", user.Username, optionIdx, pollID))
	} else {
		a.mmClient.CreatePost(channelID, fmt.Sprintf("Ваш голос за вариант %d в голосовании `%s` принят.", optionIdx, pollID))
	}
}

func (a *App) handleResults(channelID string, args []string) {
	if len(args) < 1 {
		a.mmClient.CreatePost(channelID, "Ошибка: Укажите ID голосования. Используйте: results [ID голосования]")
		return
	}
	
	pollID := args[0]
	
	results, err := a.repository.GetPollResults(pollID)
	if err != nil {
		a.mmClient.CreatePost(channelID, fmt.Sprintf("Ошибка: Голосование с ID `%s` не найдено.", pollID))
		return
	}
	
	message := formatResultsMessage(results)
	
	a.mmClient.CreatePost(channelID, message)
}

func (a *App) handleFinishPoll(userID, channelID string, args []string) {
	if len(args) < 1 {
		a.mmClient.CreatePost(channelID, "Ошибка: Укажите ID голосования. Используйте: finish [ID голосования]")
		return
	}
	
	pollID := args[0]
	
	poll, err := a.repository.GetPoll(pollID)
	if err != nil {
		a.mmClient.CreatePost(channelID, fmt.Sprintf("Ошибка: Голосование с ID `%s` не найдено.", pollID))
		return
	}
	
	if poll.CreatorID != userID {
		a.mmClient.CreatePost(channelID, "Ошибка: Только создатель голосования может его завершить.")
		return
	}
	
	if poll.IsFinished {
		a.mmClient.CreatePost(channelID, "Голосование уже завершено.")
		return
	}
	
	poll.IsFinished = true
	poll.FinishedAt = time.Now()
	
	err = a.repository.UpdatePoll(poll)
	if err != nil {
		a.logger.WithError(err).Error("Failed to update poll")
		a.mmClient.CreatePost(channelID, "Ошибка при обновлении голосования.")
		return
	}
	
	results, err := a.repository.GetPollResults(pollID)
	if err != nil {
		a.logger.WithError(err).Error("Failed to get poll results")
		a.mmClient.CreatePost(channelID, "Ошибка при получении результатов голосования.")
		return
	}
	
	message := formatResultsMessage(results)
	message = "### Голосование завершено!\n" + message
	
	a.mmClient.CreatePost(channelID, message)
}

func (a *App) handleDeletePoll(userID, channelID string, args []string) {
	if len(args) < 1 {
		a.mmClient.CreatePost(channelID, "Ошибка: Укажите ID голосования. Используйте: delete [ID голосования]")
		return
	}
	
	pollID := args[0]
	
	poll, err := a.repository.GetPoll(pollID)
	if err != nil {
		a.mmClient.CreatePost(channelID, fmt.Sprintf("Ошибка: Голосование с ID `%s` не найдено.", pollID))
		return
	}
	
	if poll.CreatorID != userID {
		a.mmClient.CreatePost(channelID, "Ошибка: Только создатель голосования может его удалить.")
		return
	}
	
	err = a.repository.DeletePoll(pollID)
	if err != nil {
		a.logger.WithError(err).Error("Failed to delete poll")
		a.mmClient.CreatePost(channelID, "Ошибка при удалении голосования.")
		return
	}
	
	a.mmClient.CreatePost(channelID, fmt.Sprintf("Голосование с ID `%s` успешно удалено.", pollID))
}


func splitQuoted(s string) []string {
	var result []string
	var current string
	inQuotes := false
	
	for _, char := range s {
		if char == '"' {
			inQuotes = !inQuotes
			if !inQuotes && current != "" {
				result = append(result, current)
				current = ""
			}
		} else if inQuotes {
			current += string(char)
		}
	}
	
	if current != "" {
		result = append(result, current)
	}
	
	return result
}

func formatPollMessage(poll models.Poll, results *models.PollResults) string {
	message := fmt.Sprintf("### %s\n", poll.Title)
	message += fmt.Sprintf("**ID голосования**: `%s`\n\n", poll.ID)
	
	for i, option := range poll.Options {
		count := 0
		if results != nil {
			count = results.Results[option]
		}
		
		message += fmt.Sprintf("%d. %s", i+1, option)
		if results != nil {
			message += fmt.Sprintf(" (%d голосов)", count)
		}
		message += "\n"
	}
	
	message += "\nДля голосования отправьте: `vote " + poll.ID + " [номер варианта]`"
	
	if poll.IsFinished {
		message += "\n\n**Голосование завершено!**"
	}
	
	return message
}

func formatResultsMessage(results models.PollResults) string {
	message := fmt.Sprintf("### Результаты голосования: %s\n", results.Poll.Title)
	message += fmt.Sprintf("**ID голосования**: `%s`\n\n", results.Poll.ID)
	
	totalVotes := 0
	for _, count := range results.Results {
		totalVotes += count
	}
	
	for i, option := range results.Poll.Options {
		count := results.Results[option]
		percentage := 0.0
		if totalVotes > 0 {
			percentage = float64(count) / float64(totalVotes) * 100
		}
		
		message += fmt.Sprintf("%d. **%s**: %d голосов (%.1f%%)\n", i+1, option, count, percentage)
	}
	
	message += fmt.Sprintf("\n**Всего голосов**: %d", totalVotes)
	
	if results.Poll.IsFinished {
		message += "\n**Статус**: Завершено"
		if !results.Poll.FinishedAt.IsZero() {
			message += fmt.Sprintf(" (%s)", results.Poll.FinishedAt.Format("02.01.2006 15:04:05"))
		}
	} else {
		message += "\n**Статус**: Активно"
	}
	
	return message
}