package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	traq "github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	payload "github.com/traPtitech/traq-ws-bot/payload"
)

// "github.com/mathsuky/BOT_remind/cache"
// "github.com/mathsuky/BOT_remind/query"
// "github.com/mathsuky/BOT_remind/transport"
// "github.com/hasura/go-graphql-client"

// const githubGraphQLEndpoint = "https://api.github.com/graphql"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}

	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: os.Getenv("TRAQ_BOT_ACCESS_TOKEN"),
	})
	if err != nil {
		panic(err)
	}

	// content := `!{"type":"user","raw":"@masky5859","id":"1eea935c-0d3c-411b-a565-1b09565237f4"} is :lol_Slander_Max_Baka:`
	var content string
	bot.OnMessageCreated(func(p *payload.MessageCreated) {
		log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)
		content = p.Message.Text
		parts := strings.Split(content, " ")
		user := p.Message.User
		log.Println(user)
		log.Println(parts)
		if parts[1] == "/hello" {
			content = "Hello, world!"
		} else if parts[1] == "/baka" {
			content = "Baka is " + fmt.Sprintf(`!{"type":"user","raw":"@%s","id":"%s"}`, user.Name, user.ID) + "!"
		} else {
			content = "Unknown command."
		}
		_, _, err := bot.API().
			MessageApi.
			PostMessage(context.Background(), p.Message.ChannelID).
			PostMessageRequest(traq.PostMessageRequest{
				Content: content,
			}).
			Execute()
		if err != nil {
			log.Println(err)
		}
	})

	if err := bot.Start(); err != nil {
		panic(err)
	}

	// githubToken := os.Getenv("GITHUB_TOKEN_CLASSIC")
	// httpClient := &http.Client{
	// 	Transport: &transport.Transport{
	// 		Token: githubToken,
	// 	},
	// }
	// client := graphql.NewClient(githubGraphQLEndpoint, httpClient)

	// // キャッシュから情報を取得
	// baseInfo, err := cache.LoadCache()
	// var projectId string
	// var issuesDict map[int]string
	// var fieldsDict map[string]graphql.ID
	// if err != nil {
	// 	projectId, issuesDict, fieldsDict, err = cache.MakeCache(client)
	// 	if err != nil {
	// 		log.Fatalf("failed to make cache: %v", err)
	// 	}
	// 	err = cache.SaveCache(projectId, issuesDict, fieldsDict)
	// 	if err != nil {
	// 		log.Fatalf("failed to save cache: %v", err)
	// 	}
	// }

	// // 以下はミューテーションの例
	// var targetIssueId int
	// fmt.Printf("Enter the issue number: ")
	// _, err = fmt.Scanf("%d", &targetIssueId)
	// if err != nil {
	// 	log.Fatalf("failed to read issue number: %v", err)
	// }

	// itemId, ok := baseInfo.IssuesDict[targetIssueId]
	// if !ok {
	// 	projectId, issuesDict, fieldsDict, err = cache.MakeCache(client)
	// 	if err != nil {
	// 		log.Fatalf("failed to make cache: %v", err)
	// 	}
	// 	err = cache.SaveCache(projectId, issuesDict, fieldsDict)
	// 	if err != nil {
	// 		log.Fatalf("failed to save cache: %v", err)
	// 	}
	// 	itemId, ok = issuesDict[targetIssueId]
	// 	if !ok {
	// 		log.Fatalf("issue number %d is not found", targetIssueId)
	// 	}
	// }

	// input := query.UpdateProjectV2ItemFieldValueInput{
	// 	ItemID:    graphql.ID(itemId),
	// 	ProjectID: "PVT_kwHOBZSipc4AuISm",
	// 	FieldID:   "PVTF_lAHOBZSipc4AuISmzgkxryw",
	// 	Value: struct {
	// 		Date graphql.String `json:"date"`
	// 	}{
	// 		Date: graphql.String("2025-05-02"),
	// 	},
	// }
	// var m query.Mutation
	// log.Printf("Executing mutation with input: %+v\n", input)
	// err = client.Mutate(context.Background(), &m, map[string]interface{}{
	// 	"input": input,
	// })
	// if err != nil {
	// 	log.Fatalf("failed to execute mutation: %v", err)
	// }

	// // 結果を出力
	// log.Printf("Updated project item ID: %s\n", m.UpdateProjectV2ItemFieldValue.ProjectV2Item.ID)
}
