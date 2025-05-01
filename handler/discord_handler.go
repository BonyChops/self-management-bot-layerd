package handler

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"self-management-bot/service"
	"strconv"
	"strings"
	"time"
)

var resetAllConfirm = make(map[string]time.Time)

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
	case strings.HasPrefix(content, "!done "):
		HandleComplete(s, m, content)
	case strings.HasPrefix(content, "!delete"):
		HandleDelete(s, m, content)
	case strings.HasPrefix(content, "!chat"):
		HandleChat(s, m, content)
	case strings.HasPrefix(content, "!reset"):
		HandleReset(s, m)
	case strings.HasPrefix(content, "!confirm reset"):
		HandleConfirm(s, m)
	}
}

func HandleAdd(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	title := strings.TrimPrefix(content, "!add ")
	if len(title) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```⚠️ タスク内容を追加してください```", m.Author.ID))
		return
	}
	err := service.AddTaskService(m.Author.ID, title)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```❌ タスク登録失敗```", m.Author.ID))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```⭕️ タスク追加: %s```", m.Author.ID, title))
}

func HandleList(s *discordgo.Session, m *discordgo.MessageCreate) {
	tasks, err := service.GetTaskService(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n```❌ タスク取得失敗```")
		return
	}
	if len(tasks) == 0 {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n```📭 タスクが登録されていません```")
		return
	}
	var msg strings.Builder
	msg.WriteString("今日のTodoです！\n")
	msg.WriteString("```")
	for i, task := range tasks {
		msg.WriteString(fmt.Sprintf("⌛️ [%02d] %s\n", i, task.Title))
	}
	msg.WriteString("```")
	s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n"+msg.String())
}

func HandleComplete(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	arg := strings.TrimPrefix(content, "!done ")
	DoneTaskNumber, err := strconv.Atoi(arg)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n```❌ 数字を指定してください```")
		return
	}
	err = service.CompleteTaskService(m.Author.ID, DoneTaskNumber)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n```❌ "+err.Error()+"```")
		return
	}
	// 完了 + 残タスク表示
	tasks, err := service.GetTaskService(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n✅ タスク完了！\n⚠️ 残りのタスク取得に失敗しました", m.Author.ID))
		return
	}
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("<@%s>\n```✅ タスク完了！お疲れ様です！\n", m.Author.ID))
	if len(tasks) == 0 {
		msg.WriteString("\n🎉 もう残ってるタスクはありません！今日もよく頑張った！```")
	} else {
		msg.WriteString("\n📝 残りのタスク:\n")
		for i, task := range tasks {
			msg.WriteString(fmt.Sprintf("⌛️ [%02d] %s\n", i, task.Title))
		}
		msg.WriteString("```")
	}
	s.ChannelMessageSend(m.ChannelID, msg.String())
}

func HandleDelete(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	arg := strings.TrimPrefix(content, "!delete ")
	DeleteNumber, err := strconv.Atoi(arg)
	// 入力バリデーション
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n```❌ 数字を指定してください```")
		return
	}
	err = service.DeleteTaskService(m.Author.ID, DeleteNumber)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "<@"+m.Author.ID+">\n```❌ "+err.Error()+"```")
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```⭕️ タスク削除しました```", m.Author.ID))
}

func HandleChat(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	arg := strings.TrimPrefix(content, "!chat ")
	if len(strings.TrimSpace(arg)) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```❌ メッセージを入力してください```", m.Author.ID))
		return
	}
	s.ChannelTyping(m.ChannelID) // 入力中表示
	reply, err := service.ChatWithContext(m.Author.ID, content)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```%s```", m.Author.ID, err.Error()))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n%s\n", m.Author.ID, reply))
}

func HandleReset(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, "!reset all") {
		resetAllConfirm[m.Author.ID] = time.Now().Add(10 * time.Minute) // 10分先まで有効
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
			"<@%s>\n```⚠️ 本当に全タスク（過去含む）を削除しますか？\n削除するには '!confirm reset' と入力してください。```",
			m.Author.ID,
		))
		return
	}
	// !reset（今日のみ削除）
	count, err := service.ResetTodayTasks(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```❌ 今日のリセット失敗: %s```", m.Author.ID, err.Error()))
		return
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```✅ 今日のタスクを %d 件削除しました```", m.Author.ID, count))
}
func HandleConfirm(s *discordgo.Session, m *discordgo.MessageCreate) {
	// 期限切れのリセットリクエストか確認
	expiry, ok := resetAllConfirm[m.Author.ID]
	if !ok || time.Now().After(expiry) {
		delete(resetAllConfirm, m.Author.ID)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```⚠️ '!reset all' の確認時間が切れました。再度実行してください。```", m.Author.ID))
		return
	}
	// 全てのタスクを削除
	count, err := service.ResetAllTasks(m.Author.ID)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```❌ 全削除に失敗しました: %s```", m.Author.ID, err.Error()))
		return
	}
	// 削除要求リストから削除
	delete(resetAllConfirm, m.Author.ID)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s>\n```✅ 全タスクを %d 件削除しました```", m.Author.ID, count))
}
