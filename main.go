package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hasura/go-graphql-client"
	"github.com/joho/godotenv"
)

const (
	githubGraphQLEndpoint = "https://api.github.com/graphql"
	cacheFilePath         = "cache.json"
)

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

type GetProjectBaseInfoQuery struct {
	User struct {
		ProjectV2 struct {
			Id    string
			Items struct {
				Nodes []struct {
					Id      string
					Content struct {
						Issue struct {
							Number int
						} `graphql:"... on Issue"`
					}
				}
			} `graphql:"items(first: 100)"`
			Fields struct {
				Nodes []struct {
					ProjectV2Field struct {
						Id   string
						Name string
					} `graphql:"... on ProjectV2Field"`
				}
			} `graphql:"fields(first: 100)"`
		} `graphql:"projectV2(number: $projectNumber)"`
	} `graphql:"user(login: $user)"`
}

type cacheData struct {
	ID         string
	IssuesDict map[int]string
	FieldsDict map[string]graphql.ID
}

func loadCache() (cacheData, error) {
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return cacheData{}, err
	}

	var cache cacheData
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return cacheData{}, err
	}

	return cache, nil
}

func makeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, error) {
	var info GetProjectBaseInfoQuery
	// キャッシュがない場合はクエリを実行してキャッシュを保存
	err := client.Query(context.Background(), &info, map[string]interface{}{
		"projectNumber": graphql.Int(3),
		"user":          graphql.String("mathsuky"),
	})
	if err != nil {
		return "", nil, nil, err
	}

	projectId := info.User.ProjectV2.Id
	issuesDict := make(map[int]string)
	for _, item := range info.User.ProjectV2.Items.Nodes {
		issuesDict[item.Content.Issue.Number] = item.Id
	}
	fieldsDict := make(map[string]graphql.ID)
	for _, field := range info.User.ProjectV2.Fields.Nodes {
		fieldsDict[field.ProjectV2Field.Name] = graphql.ID(field.ProjectV2Field.Id)
	}

	return projectId, issuesDict, fieldsDict, nil
}

func saveCache(id string, dic1 map[int]string, dic2 map[string]graphql.ID) error {
	cacheData := cacheData{
		ID:         id,
		IssuesDict: dic1,
		FieldsDict: dic2,
	}

	os.Remove(cacheFilePath)

	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	err = os.WriteFile(cacheFilePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}

	githubToken := os.Getenv("GITHUB_TOKEN_CLASSIC")
	httpClient := &http.Client{
		Transport: &transport{
			token: githubToken,
		},
	}
	client := graphql.NewClient(githubGraphQLEndpoint, httpClient)

	// キャッシュから情報を取得
	baseInfo, err := loadCache()
	var projectId string
	var issuesDict map[int]string
	var fieldsDict map[string]graphql.ID
	if err != nil {
		projectId, issuesDict, fieldsDict, err = makeCache(client)
		if err != nil {
			log.Fatalf("failed to make cache: %v", err)
		}
		err = saveCache(projectId, issuesDict, fieldsDict)
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
		projectId, issuesDict, fieldsDict, err = makeCache(client)
		if err != nil {
			log.Fatalf("failed to make cache: %v", err)
		}
		err = saveCache(projectId, issuesDict, fieldsDict)
		if err != nil {
			log.Fatalf("failed to save cache: %v", err)
		}
		itemId, ok = issuesDict[targetIssueId]
		if !ok {
			log.Fatalf("issue number %d is not found", targetIssueId)
		}
	}

	input := UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(itemId),
		ProjectID: "PVT_kwHOBZSipc4AuISm",
		FieldID:   "PVTF_lAHOBZSipc4AuISmzgkxryw",
		Value: struct {
			Date graphql.String `json:"date"`
		}{
			Date: graphql.String("2025-05-02"),
		},
	}
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
