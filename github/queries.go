package github

import (
	"context"
	"fmt"

	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/cache"
	"github.com/mathsuky/BOT_remind/query"
)

func UpdateDeadline(client *graphql.Client, date string, targetIssueId int) (string, error) {
	// APIを叩くための基本情報を取得
	projectId, issuesDict, fieldsDict, err := cache.LoadOrMakeCache(client)
	if err != nil {
		return "キャッシュの読み込みまたは作成に失敗しました。", err
	}
	goalItemId, goalFieldId, err := CheckIssueAndField(client, issuesDict, fieldsDict, targetIssueId, "目標")
	if err != nil {
		return "issueが紐づけられていないか，「目標」フィールドが存在しませんでした。", err
	}
	startItemId, startFieldId, err := CheckIssueAndField(client, issuesDict, fieldsDict, targetIssueId, "目標開始日")
	if err != nil {
		return "issueが紐づけられていないか，「目標開始日」フィールドが存在しませんでした。", err
	}

	// APIを叩くための変数を設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "タイムゾーンの取得に失敗しました。", err
	}
	nowDate := time.Now().In(jst).Format("2006-01-02")

	input1 := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(goalItemId),
		ProjectID: graphql.ID(projectId),
		FieldID:   graphql.ID(goalFieldId),
		Value: struct {
			Date graphql.String `json:"date"`
		}{
			Date: graphql.String(date),
		},
	}
	input2 := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(startItemId),
		ProjectID: graphql.ID(projectId),
		FieldID:   graphql.ID(startFieldId),
		Value: struct {
			Date graphql.String `json:"date"`
		}{
			Date: graphql.String(nowDate),
		},
	}
	vars := map[string]interface{}{
		"input1": input1,
		"input2": input2,
	}

	// ミューテーションの実行
	var mutation query.UpdateTwoFieldsMutation
	err = client.Mutate(context.Background(), &mutation, vars)
	if err != nil {
		return "ミューテーションの実行に失敗しました。", err
	}
	return "期日が正常に設定されました。", nil
}

func CheckIssueAndField(client *graphql.Client, issuesDict map[int]string, fieldsDict map[string]graphql.ID, targetIssueId int, fieldKey string) (string, graphql.ID, error) {
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
