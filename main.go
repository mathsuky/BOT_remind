package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
)

// GitHub APIのエンドポイント
const githubGraphQLEndpoint = "https://api.github.com/graphql"

// 特定のプロジェクトのIssueを取得するためのクエリ用構造体
type ProjectIssuesQuery struct {
	User struct {
		ProjectV2 struct {
			Items struct {
				Nodes []struct {
					Content struct {
						Issue struct {
							Title  string
							Number int
						} `graphql:"... on Issue"`
					}
					Kijitu struct {
						ProjectV2ItemFieldDateValue struct {
							Date string
						} `graphql:"... on ProjectV2ItemFieldDateValue"`
					} `graphql:"fieldValueByName(name: \"kijitu\")"`
				}
			} `graphql:"items(first: 100)"`
		} `graphql:"projectV2(number: $projectNumber)"`
	} `graphql:"user(login: $user)"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}
	// GitHubのPersonal Access Token
	githubToken := os.Getenv("GITHUB_TOKEN_CLASSIC")

	// HTTPクライアントを作成し、Authorizationヘッダーを設定
	httpClient := &http.Client{
		Transport: &transport{
			token: githubToken,
		},
	}

	// GraphQLクライアントを初期化
	client := graphql.NewClient(githubGraphQLEndpoint, httpClient)

	// クエリ結果を格納する変数
	var query ProjectIssuesQuery

	// クエリ変数を設定
	variables := map[string]interface{}{
		"projectNumber": graphql.Int(3),             // プロジェクト番号を指定
		"user":          graphql.String("mathsuky"), // ユーザー名を指定
	}

	// クエリ実行
	err = client.Query(context.Background(), &query, variables)
	if err != nil {
		log.Fatalf("failed to execute query: %v", err)
	}

	// 結果を出力
	fmt.Println("Issues in the Project:")
	for _, item := range query.User.ProjectV2.Items.Nodes {
		fmt.Printf("Issue Number: %d\nTitle: %s\nKijitu Date: %s\n\n", item.Content.Issue.Number, item.Content.Issue.Title, item.Kijitu.ProjectV2ItemFieldDateValue.Date)
	}
}

// HTTPクライアント用のトランスポート構造体
type transport struct {
	token string
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultTransport.RoundTrip(req)
}
