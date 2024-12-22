package bot

import (
	"context"
	"log"

	"github.com/hasura/go-graphql-client"
	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

func Start(client *graphql.Client, accessToken string) error {
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: accessToken,
	})
	if err != nil {
		return err
	}

	// メッセージハンドラを登録
	bot.OnMessageCreated(func(p *payload.MessageCreated) {
		response, err := HandleMessage(client, p)
		if err != nil {
			log.Println(err)
		}
		// メッセージ送信
		_, _, err = bot.API().MessageApi.PostMessage(context.Background(), p.Message.ChannelID).
			PostMessageRequest(traq.PostMessageRequest{Content: response}).Execute()
		if err != nil {
			log.Println(err)
		}
	})

	// Bot の起動
	return bot.Start()
}
