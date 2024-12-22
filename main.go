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

func getClient() (*graphql.Client, error) {
	githubToken := os.Getenv("GITHUB_TOKEN_CLASSIC")
	if githubToken == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN_CLASSIC is not set")
	}
	httpClient := &http.Client{
		Transport: &transport.Transport{
			Token: githubToken,
		},
	}
	client := graphql.NewClient(githubGraphQLEndpoint, httpClient)
	return client, nil
}

func loadOrMakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, error) {
	baseInfo, err := cache.LoadCache()
	if err == nil {
		return baseInfo.ID, baseInfo.IssuesDict, baseInfo.FieldsDict, nil
	}

	projectId, issuesDict, fieldsDict, err := cache.MakeCache(client)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to make cache: %v", err)
	}
	err = cache.SaveCache(projectId, issuesDict, fieldsDict)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to save cache: %v", err)
	}
	return projectId, issuesDict, fieldsDict, nil
}

func updateDeadline(client *graphql.Client, date string, targetIssueId int) (string, error) {
	projectId, issuesDict, fieldsDict, err := loadOrMakeCache(client)
	if err != nil {
		return "キャッシュの読み込みまたは作成に失敗しました。", err
	}

	itemId, fieldId, err := checkIssueAndField(client, issuesDict, fieldsDict, targetIssueId, "kijitu")
	if err != nil {
		return "issueが紐づけられていないか，期日を記入するフィールドが存在しませんでした。", err
	}

	input := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(itemId),
		ProjectID: graphql.ID(projectId),
		FieldID:   graphql.ID(fieldId),
		Value: struct {
			Date graphql.String `json:"date"`
		}{
			Date: graphql.String(date),
		},
	}
	var m query.UpdateProjectV2ItemFieldValue
	log.Printf("Executing mutation with input: %+v\n", input)
	err = client.Mutate(context.Background(), &m, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return "ミューテーションの実行に失敗しました。", fmt.Errorf("failed to execute mutation: %v", err)
	}
	return "期日が正常に設定されました。", nil
}

func checkIssueAndField(client *graphql.Client, issuesDict map[int]string, fieldsDict map[string]graphql.ID, targetIssueId int, fieldKey string) (string, graphql.ID, error) {
	itemId, ok := issuesDict[targetIssueId]
	fieldId, ok2 := fieldsDict[fieldKey]
	if !ok || !ok2 {
		_, issuesDict, fieldsDict, err := cache.MakeCache(client)
		if err != nil {
			return "", "", fmt.Errorf("failed to make cache: %v", err)
		}
		itemId, ok = issuesDict[targetIssueId]
		fieldId, ok2 = fieldsDict[fieldKey]
		if !ok || !ok2 {
			return "", "", fmt.Errorf("issue or field not found")
		}
	}
	return itemId, fieldId, nil
}

func handleMessage(client *graphql.Client, p *payload.MessageCreated) (string, error) {
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
		// required: @BOT_remind deadline <date> <issue number>
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
		errorString, err := updateDeadline(client, date, targetIssueId)
		if err != nil {
			return errorString, err
		}
		return errorString, nil
	}
	return "正しいコマンドを入力してください。", nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}

	client, err := getClient()
	if err != nil {
		log.Fatalf("failed to get client: %v", err)
	}
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: os.Getenv("TRAQ_BOT_ACCESS_TOKEN"),
	})
	if err != nil {
		panic(err)
	}

	var content string
	bot.OnMessageCreated(func(p *payload.MessageCreated) {
		content, err = handleMessage(client, p)
		if err != nil {
			log.Println(err)
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
