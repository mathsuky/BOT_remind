package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/github"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func HandleMessage(client *graphql.Client, p *payload.MessageCreated) (string, error) {
	log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)
	parts := strings.Split(p.Message.Text, " ")

	if len(parts) < 2 {
		return "有効なコマンドを入力してください。", nil
	}

	switch parts[1] {
	case "/hello":
		return "Hello, world!", nil
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
	default:
		return "正しいコマンドを入力してください。", nil
	}
}
