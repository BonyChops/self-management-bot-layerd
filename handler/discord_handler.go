package handler

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"self-management-bot/service"
	"strings"
)

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.TrimSpace(m.Content)

	switch {
	case strings.HasPrefix(content, "!add "):
		HandleAdd(s, m, content)
	case strings.HasPrefix(content, "!list"):
		HandleList(s, m)
	}
}

func HandleAdd(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	title := strings.TrimPrefix(content, "!add ")
	err := service.AddTaskService(m.Author.ID, title)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```❌ タスク登録失敗```", m.Author.ID))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```✅ タスク追加: %s```", m.Author.ID, title))
}

func HandleList(s *discordgo.Session, m *discordgo.MessageCreate) {
	tasks, err := service.GetTaskService(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n "+"❌ タスク取得失敗")
		return
	}
	if len(tasks) == 0 {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n "+"📭 タスクが登録されていません")
		return
	}
	var msg strings.Builder
	msg.WriteString("今日のTodoです！\n")
	msg.WriteString("```")
	for i, task := range tasks {
		status := "⌛️"
		if task.Status == "Completed" {
			status = "✅"
		}
		msg.WriteString(fmt.Sprintf("%s [%02d] %s\n", status, i, task.Title))
	}
	msg.WriteString("```")
	s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n"+msg.String())

}
