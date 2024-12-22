package github

import (
	"context"
	"fmt"
	"log"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/cache"
	"github.com/mathsuky/BOT_remind/query"
)

// UpdateDeadline は指定された issue に期日を設定する
func UpdateDeadline(client *graphql.Client, date string, targetIssueId int) (string, error) {
	// キャッシュの読み込みまたは作成
	projectId, issuesDict, fieldsDict, err := cache.loadOrMakeCache(client)
	if err != nil {
		return "キャッシュの読み込みまたは作成に失敗しました。", err
	}

	// issue とフィールドの確認
	itemId, fieldId, err := checkIssueAndField(client, issuesDict, fieldsDict, targetIssueId, "kijitu")
	if err != nil {
		return "issueが紐づけられていないか，期日を記入するフィールドが存在しませんでした。", err
	}

	// ミューテーションの入力
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

	var mutation query.UpdateProjectV2ItemFieldValue
	log.Printf("Executing mutation with input: %+v\n", input)
	err = client.Mutate(context.Background(), &mutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return "ミューテーションの実行に失敗しました。", fmt.Errorf("failed to execute mutation: %v", err)
	}

	return "期日が正常に設定されました。", nil
}

// checkIssueAndField は指定された issue ID とフィールドキーがキャッシュに存在するか確認します
func checkIssueAndField(client *graphql.Client, issuesDict map[int]string, fieldsDict map[string]graphql.ID, targetIssueId int, fieldKey string) (string, graphql.ID, error) {
	itemId, ok := issuesDict[targetIssueId]
	fieldId, ok2 := fieldsDict[fieldKey]
	if !ok || !ok2 {
		// キャッシュを再作成
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
