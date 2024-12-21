package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
)

// GitHub APIのエンドポイント
const githubGraphQLEndpoint = "https://api.github.com/graphql"

// ミューテーション用の構造体
type UpdateProjectV2ItemFieldValueInput struct {
	ItemID    graphql.ID `json:"itemId"`
	ProjectID graphql.ID `json:"projectId"`
	FieldID   graphql.ID `json:"fieldId"`
	Value     struct {
		Date graphql.String `json:"date"`
	} `json:"value"`
}

type Mutation struct {
	UpdateProjectV2ItemFieldValue struct {
		ProjectV2Item struct {
			ID graphql.String `graphql:"id"`
		} `graphql:"projectV2Item"`
	} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}
	// GitHubのPersonal Access Token
	githubToken := os.Getenv("GITHUB_TOKEN_CLASSIC")

	// HTTPクライアントを作成し、カスタムトランスポートを設定
	httpClient := &http.Client{
		Transport: &transport{
			token: githubToken,
		},
	}

	// GraphQLクライアントを初期化
	client := graphql.NewClient(githubGraphQLEndpoint, httpClient)

	// ミューテーション用の入力データ
	input := UpdateProjectV2ItemFieldValueInput{
		ItemID:    "PVTI_lAHOBZSipc4AuISmzgVrKpU",
		ProjectID: "PVT_kwHOBZSipc4AuISm",
		FieldID:   "PVTF_lAHOBZSipc4AuISmzgkxryw",
		Value: struct {
			Date graphql.String `json:"date"`
		}{
			Date: graphql.String("2025-05-01"),
		},
	}

	// ミューテーションを実行
	var m Mutation
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

// HTTPクライアント用のトランスポート構造体（トークン設定）
type transport struct {
	token string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultTransport.RoundTrip(req)
}
