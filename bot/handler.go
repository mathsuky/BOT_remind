package bot

import (
	"fmt"
	"log"
	"strings"

	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/github"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func HandleMessage(client *graphql.Client, p *payload.MessageCreated) (string, error) {
	log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)
	log.Println("User: " + p.Message.User.Name)
	parts := strings.Split(p.Message.Text, " ")

	if len(parts) < 2 {
		return "有効なコマンドを入力してください。", nil
	}

	switch parts[1] {
	case "/hello":
		jst, err := time.LoadLocation("Asia/Tokyo")
		if err != nil {
			return "タイムゾーンの取得に失敗しました。", err
		}
		now := time.Now().In(jst)
		date := now.Format("2006-01-02")
		return "こんにちは！今日は" + date + "です。", nil
	case "/deadline":
		if len(parts) < 4 {
			return "十分な引数を提供してください。", nil
		}
		date := parts[2]
		issueID := 0
		if _, err := fmt.Sscanf(parts[3], "%d", &issueID); err != nil {
			return "有効な issue ID を入力してください。", nil
		}
		return github.UpdateDeadline(client, date, issueID)
	case "/assign": // TODO:匹数に不適切なgithubユーザー名が入力された場合の処理
		if len(parts) < 4 {
			return "十分な引数を提供してください。", nil
		}
		issueID := 0
		if _, err := fmt.Sscanf(parts[2], "%d", &issueID); err != nil {
			return "有効な issue ID を入力してください。", nil
		}
		return github.UpdateAssigner(client, issueID, p.Message.User.Name, parts[3])
	default:
		return "正しいコマンドを入力してください。", nil
	}
}
