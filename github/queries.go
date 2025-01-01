package github

import (
	"context"
	"fmt"
	"log"

	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/cache"
	"github.com/mathsuky/BOT_remind/query"
)

func UpdateDeadline(client *graphql.Client, date string, targetIssueId int) (string, error) {
	// APIを叩くための基本情報を取得
	projectId, issuesDict, fieldsDict, fieldsTypeDict, err := cache.LoadOrMakeCache(client)
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
	log.Println(fieldsTypeDict)

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
		Value: query.FieldValue{
			"date": date, // 目標フィールドがDate型の場合
		},
	}
	input2 := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(startItemId),
		ProjectID: graphql.ID(projectId),
		FieldID:   graphql.ID(startFieldId),
		Value: query.FieldValue{
			"date": nowDate, // 目標開始日フィールドがDate型の場合
		},
	}

	vars := map[string]interface{}{
		"input1": input1,
		"input2": input2,
	}

	// ミューテーションの実行
	var mutation query.UpdateRelatedIssueDeadlineMutation
	err = client.Mutate(context.Background(), &mutation, vars)
	if err != nil {
		return "ミューテーションの実行に失敗しました。", err
	}
	return "期日が正常に設定されました。", nil
}

func UpdateAssigner(client *graphql.Client, targetIssueId int, tId string, gId string) (string, error) {
	// APIを叩くための基本情報を取得
	projectId, issuesDict, fieldsDict, fieldsTypeDict, err := cache.LoadOrMakeCache(client)
	if err != nil {
		return "キャッシュの読み込みまたは作成に失敗しました。", err
	}
	traqIDItemId, traqIDFieldId, err := CheckIssueAndField(client, issuesDict, fieldsDict, targetIssueId, "traQID")
	if err != nil {
		return "issueが紐づけられていないか，「traQID」フィールドが存在しませんでした。", err
	}
	log.Println(fieldsTypeDict)
	log.Println(traqIDItemId)

	// APIを叩くための変数を設定
	updateInput := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(traqIDItemId),
		ProjectID: graphql.ID(projectId),
		FieldID:   graphql.ID(traqIDFieldId),
		Value: query.FieldValue{
			"text": tId,
		},
	}
	updateVars := map[string]interface{}{
		"input": updateInput,
	}

	// traQIDフィールドを更新するmutationを実行
	var updateMutation query.UpdateProjectV2ItemFieldValue
	err = client.Mutate(context.Background(), &updateMutation, updateVars)
	if err != nil {
		return "traQIDの更新に失敗しました", err
	}

	// issue ID の取得
	var issueQuery query.GetIssueIdFromRepositoryQuery
	err = client.Query(context.Background(), &issueQuery, map[string]interface{}{
		"owner":       graphql.String("mathsuky"), //TODO: この辺を環境変数に？
		"repo":        graphql.String("BOT_remind"),
		"issueNumber": graphql.Int(targetIssueId),
	})
	if err != nil {
		return "issue ID の取得に失敗しました。", err
	}
	issueId := issueQuery.Repository.Issue.Id
	// ユーザー ID の取得
	var assigneeQuery query.GetUserIdQuery
	err = client.Query(context.Background(), &assigneeQuery, map[string]interface{}{
		"login": graphql.String(gId),
	})
	if err != nil {
		return "ユーザー ID の取得に失敗しました。", err
	}
	assigneeId := assigneeQuery.User.Id

	// APIを叩くための変数を設定
	assignInput := query.AddAssigneesToAssignableInput{
		AssignableId: graphql.ID(issueId),
		AssigneeIds:  []graphql.ID{graphql.ID(assigneeId)},
	}
	assignVars := map[string]interface{}{
		"input": assignInput,
	}

	// Assigneeを追加するmutationを実行
	var assignMutation query.AddAssigneeToAssignableMutation
	err = client.Mutate(context.Background(), &assignMutation, assignVars)
	if err != nil {
		return "担当者の追加に失敗しました。", err
	}
	log.Println(assignMutation.AddAssigneesToAssignable.Assignable.Issue.Title)
	log.Println("sdf")

	return "担当者が正常に設定されました。", nil
}

func CheckIssueAndField(client *graphql.Client, issuesDict map[int]string, fieldsDict map[string]graphql.ID, targetIssueId int, fieldKey string) (string, graphql.ID, error) {
	itemId, ok := issuesDict[targetIssueId]
	fieldId, ok2 := fieldsDict[fieldKey]
	if !ok || !ok2 {
		// キャッシュを再作成
		_, issuesDict, fieldsDict, _, err := cache.MakeCache(client)
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
