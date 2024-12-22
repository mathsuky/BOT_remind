package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
	"github.com/mathsuky/BOT_remind/cache"
	"github.com/mathsuky/BOT_remind/query"
	"github.com/mathsuky/BOT_remind/transport"
	traq "github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	payload "github.com/traPtitech/traq-ws-bot/payload"
)

const githubGraphQLEndpoint = "https://api.github.com/graphql"

func setDeadline(date string, targetIssueId int) error {
	githubToken := os.Getenv("GITHUB_TOKEN_CLASSIC")
	httpClient := &http.Client{
		Transport: &transport.Transport{
			Token: githubToken,
		},
	}
	client := graphql.NewClient(githubGraphQLEndpoint, httpClient)

	baseInfo, err := cache.LoadCache()
	var projectId string
	var issuesDict map[int]string
	var fieldsDict map[string]graphql.ID
	if err != nil {
		projectId, issuesDict, fieldsDict, err = cache.MakeCache(client)
		if err != nil {
			return fmt.Errorf("failed to make cache: %v", err)
		}
		err = cache.SaveCache(projectId, issuesDict, fieldsDict)
		if err != nil {
			return fmt.Errorf("failed to save cache: %v", err)
		}
	}

	itemId, ok := baseInfo.IssuesDict[targetIssueId]
	if !ok {
		projectId, issuesDict, fieldsDict, err = cache.MakeCache(client)
		if err != nil {
			return fmt.Errorf("failed to make cache: %v", err)
		}
		err = cache.SaveCache(projectId, issuesDict, fieldsDict)
		if err != nil {
			return fmt.Errorf("failed to save cache: %v", err)
		}
		itemId, ok = issuesDict[targetIssueId]
		if !ok {
			return fmt.Errorf("issue number %d is not found", targetIssueId)
		}
	}

	input := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(itemId),
		ProjectID: "PVT_kwHOBZSipc4AuISm",
		FieldID:   "PVTF_lAHOBZSipc4AuISmzgkxryw",
		Value: struct {
			Date graphql.String `json:"date"`
		}{
			Date: graphql.String(date),
		},
	}
	var m query.Mutation
	log.Printf("Executing mutation with input: %+v\n", input)
	err = client.Mutate(context.Background(), &m, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return fmt.Errorf("failed to execute mutation: %v", err)
	}

	// 結果を出力
	log.Printf("Updated project item ID: %s\n", m.UpdateProjectV2ItemFieldValue.ProjectV2Item.ID)
	return nil
}

func handleMessage(p *payload.MessageCreated) (string, error) {
	log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)
	content := p.Message.Text
	parts := strings.Split(content, " ")
	user := p.Message.User
	log.Println(user)
	log.Println(parts)
	if len(parts) < 2 {
		return "有効なコマンドを入力してください。", nil
	}
	if parts[1] == "/hello" {
		return "Hello, world!", nil
	} else if parts[1] == "/baka" {
		return "Baka is " + fmt.Sprintf(`!{"type":"user","raw":"@%s","id":"%s"}`, user.Name, user.ID) + "!", nil
	} else if parts[1] == "/deadline" {
		// requierd: @BOT_remind deadline <date> <issue number>
		if len(parts) < 4 {
			return "十分な引数を提供してください。", nil
		}
		date := parts[2]
		targetIssueId := 0
		_, err := fmt.Sscanf(parts[3], "%d", &targetIssueId)
		log.Println(targetIssueId)
		log.Println(date)
		if err != nil {
			return "有効なコマンドを入力してください。", nil
		}
		err = setDeadline(date, targetIssueId)
		if err != nil {
			return "期日の設定に失敗しました。", err
		}
	}
	return "正しいコマンドを入力してください。", nil
}

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
		content, err = handleMessage(p)
		if err != nil {
			log.Println(err)
			return
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
}
