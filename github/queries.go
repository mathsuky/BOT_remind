package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/cache"
	"github.com/mathsuky/BOT_remind/query"
)

func UpdateDeadline(client *graphql.Client, date string, targetIssueId int) (string, error) {
	// APIを叩くための基本情報を取得
	projectId, issuesDict, fieldsDict, _, err := cache.LoadOrMakeCache(client)
	if err != nil {
		return "キャッシュの読み込みまたは作成に失敗しました。", err
	}
	goalItemId, goalFieldId, err := GetItemAndFieldId(client, issuesDict, fieldsDict, targetIssueId, "目標")
	if err != nil {
		return "issueが紐づけられていないか，「目標」フィールドが存在しませんでした。", err
	}
	startItemId, startFieldId, err := GetItemAndFieldId(client, issuesDict, fieldsDict, targetIssueId, "目標開始日")
	if err != nil {
		return "issueが紐づけられていないか，「目標開始日」フィールドが存在しませんでした。", err
	}

	// APIを叩くための変数を設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "タイムゾーンの取得に失敗しました。", err
	}
	nowDate := time.Now().In(jst).Format("2006-01-02")

	// dateが今日以降かを確認
	if date < nowDate {
		return "過去の日付は設定できません。", nil
	}

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

func UpdateAssigner(client *graphql.Client, targetIssueNum int, traqID string, githubLogin string) (string, error) {
	// APIを叩くための基本情報を取得
	projectID, issuesDict, fieldsDict, statusOptionsDict, err := cache.LoadOrMakeCache(client)
	if err != nil {
		return "キャッシュの読み込みまたは作成に失敗しました。", err
	}
	itemID, traqIDFieldID, err := GetItemAndFieldId(client, issuesDict, fieldsDict, targetIssueNum, "traQID")
	if err != nil {
		return "issueが紐づけられていないか，「traQID」フィールドが存在しませんでした。", err
	}
	log.Println(itemID)

	// APIを叩くための変数を設定
	updateInput := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(itemID),
		ProjectID: graphql.ID(projectID),
		FieldID:   graphql.ID(traqIDFieldID),
		Value: query.FieldValue{
			"text": traqID,
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
		"owner":       graphql.String(os.Getenv("REPOSITORY_OWNER")),
		"repo":        graphql.String(os.Getenv("REPOSITORY_NAME")),
		"issueNumber": graphql.Int(targetIssueNum),
	})
	if err != nil {
		return "issue ID の取得に失敗しました。", err
	}
	issueID := issueQuery.Repository.Issue.Id

	// ユーザー ID の取得
	var assigneeQuery query.GetUserIdQuery
	err = client.Query(context.Background(), &assigneeQuery, map[string]interface{}{
		"login": graphql.String(githubLogin),
	})
	if err != nil {
		return "ユーザー ID の取得に失敗しました。", err
	}
	assigneeID := assigneeQuery.User.Id

	// APIを叩くための変数を設定
	assignInput := query.AddAssigneesToAssignableInput{
		AssignableId: graphql.ID(issueID),
		AssigneeIds:  []graphql.ID{graphql.ID(assigneeID)},
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

	var statusQuery query.UpdateProjectV2ItemFieldValue
	_, statusFieldID, err := GetItemAndFieldId(client, issuesDict, fieldsDict, targetIssueNum, "Status")
	progressID := statusOptionsDict["In Progress"]
	if err != nil {
		return "issueが紐づけられていないか，「ステータス」フィールドが存在しませんでした。", err
	}
	statusUpdateInput := query.UpdateProjectV2ItemFieldValueInput{
		ItemID:    graphql.ID(itemID),
		ProjectID: graphql.ID(projectID),
		FieldID:   graphql.ID(statusFieldID),
		Value: query.FieldValue{
			"singleSelectOptionId": progressID,
		},
	}
	statusUpdateVars := map[string]interface{}{
		"input": statusUpdateInput,
	}
	err = client.Mutate(context.Background(), &statusQuery, statusUpdateVars)
	if err != nil {
		return "ステータスの更新に失敗しました。", err
	}

	return "担当者が正常に設定されました。", nil
}

// TODO:IssueDetailの定義を書く場所を考える
func Remind(client *graphql.Client) ([]query.IssueDetail, error) {
	projectNumber, err := strconv.Atoi(os.Getenv("PROJECTV2_NUMBER"))
	if err != nil {
		return nil, fmt.Errorf("failed to convert PROJECTV2_NUMBER to int: %v", err)
	}
	var tmpQuery query.GetIssueFieldsQuery
	err = client.Query(context.Background(), &tmpQuery, map[string]interface{}{
		"projectNumber": graphql.Int(projectNumber),
		"user":          graphql.String(os.Getenv("REPOSITORY_OWNER")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get issue fields: %v", err)
	}
	log.Println(tmpQuery.User.ProjectV2.Items.Nodes)
	var issues []query.IssueDetail
	for _, item := range tmpQuery.User.ProjectV2.Items.Nodes {
		// TODO: エラーハンドリングをももうちょいよくして
		if item.Deadline.ProjectV2ItemFieldDateValue.Date == "" {
			continue
		}
		deadline, err := time.Parse("2006-01-02", item.Deadline.ProjectV2ItemFieldDateValue.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse deadline: %v", err)
		}
		status := item.Status.ProjectV2ItemFieldSingleSelectValue.Name
		// 日付の差分がマイナス，0，1，3，7の場合にリマインド
		if status != "In Progress" {
			continue
		}
		daysUntilDeadline := int(time.Until(deadline).Hours() / 24)
		// TODO: この辺の条件式をもうちょいよくする
		if daysUntilDeadline == 0 || daysUntilDeadline == 1 || daysUntilDeadline == 3 || daysUntilDeadline == 7 || daysUntilDeadline < 0 {
			issues = append(issues, query.IssueDetail{IssueNum: item.Content.Issue.Number, Assignee: item.TraqID.ProjectV2ItemFieldTextValue.Text, Deadline: deadline, Status: status})
		}
	}
	log.Println(issues)
	return issues, nil
}

func GetItemAndFieldId(client *graphql.Client, issuesDict map[int]string, fieldsDict map[string]graphql.ID, targetIssueNum int, fieldKey string) (string, graphql.ID, error) {
	itemId, ok := issuesDict[targetIssueNum]
	fieldId, ok2 := fieldsDict[fieldKey]
	if !ok || !ok2 {
		// キャッシュを再作成
		_, issuesDict, fieldsDict, _, err := cache.MakeCache(client)
		if err != nil {
			return "", "", fmt.Errorf("failed to make cache: %v", err)
		}
		itemId, ok = issuesDict[targetIssueNum]
		fieldId, ok2 = fieldsDict[fieldKey]
		if !ok || !ok2 {
			return "", "", fmt.Errorf("issue or field not found")
		}
	}
	return itemId, fieldId, nil
}
