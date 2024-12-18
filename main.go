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
			Id    string
			Items struct {
				Nodes []struct {
					Id      string // ItemIdを取得するためのフィールドを追加
					Content struct {
						Issue struct {
							Title  string
							Number int
						} `graphql:"... on Issue"`
					}
					Kijitu struct {
						ProjectV2ItemFieldDateValue struct {
							Id   string
							Date string
						} `graphql:"... on ProjectV2ItemFieldDateValue"`
					} `graphql:"fieldValueByName(name: \"kijitu\")"`
				}
			} `graphql:"items(first: 100)"`
			// Fields struct {
			// 	Nodes []struct {
			// 		ProjectV2Field struct {
			// 			Id   string
			// 			Name string
			// 		} `graphql:"... on ProjectV2Field"`
			// 	}
			// } `graphql:"fields(first: 100)"`
			Field struct {
				ProjectV2Field struct {
					Id   string
					Name string
				} `graphql:"... on ProjectV2Field"`
			} `graphql:"field(name: \"kijitu\")"`
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
		"projectNumber": graphql.Int(3), // プロジェクト番号を指定
		"user":          graphql.String("mathsuky"),
	}

	// クエリ実行
	err = client.Query(context.Background(), &query, variables)
	if err != nil {
		log.Fatalf("failed to execute query: %v", err)
	}
	fmt.Println("ProjectId: ", query.User.ProjectV2.Id)
	// 結果を出力
	fmt.Println("Issues in the Project:")
	for _, item := range query.User.ProjectV2.Items.Nodes {
		fmt.Printf("Item ID: %s\nIssue Number: %d\nTitle: %s\nKijitu Date: %s\nKijitu Field ID: %s\n\n", item.Id, item.Content.Issue.Number, item.Content.Issue.Title, item.Kijitu.ProjectV2ItemFieldDateValue.Date, item.Kijitu.ProjectV2ItemFieldDateValue.Id)
	}
	fmt.Println("Fields in the Project:")
	// for _, field := range query.User.ProjectV2.Fields.Nodes {
	// 	fmt.Printf("Field ID: %s\nField Name: %s\n\n", field.ProjectV2Field.Id, field.ProjectV2Field.Name)
	// }
	fmt.Printf("Field ID: %s\nField Name: %s\n\n", query.User.ProjectV2.Field.ProjectV2Field.Id, query.User.ProjectV2.Field.ProjectV2Field.Name)
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

// gh api graphql -f query='
// mutation {
//   updateProjectV2ItemFieldValue(input: {
//     itemId: "PVTI_lAHOBZSipc4AuISmzgVltkE",
//     projectId: "PVT_kwHOBZSipc4AuISm",
//     fieldId: "PVTF_lAHOBZSipc4AuISmzgkxryw",
//     value: {
//       date: "2024-12-31"
//     }
//   }) {
//     projectV2Item {
//       id
//     }
//   }
// }
// '
