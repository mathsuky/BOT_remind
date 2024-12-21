package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mathsuky/BOT_remind/cache"
	customgraphql "github.com/mathsuky/BOT_remind/query"
	"github.com/mathsuky/BOT_remind/transport"

	graphql "github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
)

const githubGraphQLEndpoint = "https://api.github.com/graphql"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}

	githubToken := os.Getenv("GITHUB_TOKEN_CLASSIC")
	httpClient := &http.Client{
		Transport: &transport.Transport{
			Token: githubToken,
		},
	}
	client := graphql.NewClient(githubGraphQLEndpoint, httpClient)

	// キャッシュから情報を取得
	baseInfo, err := cache.LoadCache()
	var projectId string
	var issuesDict map[int]string
	var fieldsDict map[string]graphql.ID
	if err != nil {
		projectId, issuesDict, fieldsDict, err = cache.MakeCache(client)
		if err != nil {
			log.Fatalf("failed to make cache: %v", err)
		}
		err = cache.SaveCache(projectId, issuesDict, fieldsDict)
		if err != nil {
			log.Fatalf("failed to save cache: %v", err)
		}
	}

	// 以下はミューテーションの例
	var targetIssueId int
	fmt.Printf("Enter the issue number: ")
	_, err = fmt.Scanf("%d", &targetIssueId)
	if err != nil {
		log.Fatalf("failed to read issue number: %v", err)
	}

	itemId, ok := baseInfo.IssuesDict[targetIssueId]
	if !ok {
		projectId, issuesDict, fieldsDict, err = cache.MakeCache(client)
		if err != nil {
			log.Fatalf("failed to make cache: %v", err)
		}
		err = cache.SaveCache(projectId, issuesDict, fieldsDict)
		if err != nil {
			log.Fatalf("failed to save cache: %v", err)
		}
		itemId, ok = issuesDict[targetIssueId]
		if !ok {
			log.Fatalf("issue number %d is not found", targetIssueId)
		}
	}

	input := customgraphql.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(itemId),
		ProjectID: "PVT_kwHOBZSipc4AuISm",
		FieldID:   "PVTF_lAHOBZSipc4AuISmzgkxryw",
		Value: struct {
			Date graphql.String `json:"date"`
		}{
			Date: graphql.String("2025-05-02"),
		},
	}
	var m customgraphql.Mutation
	log.Printf("Executing mutation with input: %+v\n", input)
	err = client.Mutate(context.Background(), &m, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		log.Fatalf("failed to execute mutation: %v", err)
	}

	// 結果を出力
	log.Printf("Updated project item ID: %s\n", m.UpdateProjectV2ItemFieldValue.ProjectV2Item.ID)
}
