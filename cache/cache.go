package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/query"
)

// グローバルなインメモリキャッシュ
var memCache *CacheData

// キャッシュデータ構造体
type CacheData struct {
	ID                string
	IssuesDict        map[int]string
	FieldsDict        map[string]graphql.ID
	StatusOptionsDict map[string]graphql.ID
	CreatedAt         time.Time
}

// キャッシュの有効期間（24時間）
const cacheTTL = 24 * time.Hour

// LoadCache はインメモリキャッシュからデータを読み込み、有効期限をチェックします。
func LoadCache() (CacheData, error) {
	if memCache == nil {
		log.Printf("キャッシュが存在しません")
		return CacheData{}, fmt.Errorf("キャッシュが存在しません")
	}
	if time.Since(memCache.CreatedAt) > cacheTTL {
		log.Printf("キャッシュが期限切れです: 作成時刻 %v", memCache.CreatedAt)
		memCache = nil
		return CacheData{}, fmt.Errorf("キャッシュが期限切れです")
	}
	log.Printf("キャッシュをメモリから読み込みました: 作成時刻 %v", memCache.CreatedAt)
	return *memCache, nil
}

// SaveCache はキャッシュデータをインメモリに保存します。
func SaveCache(data CacheData) {
	memCache = &data
	log.Printf("キャッシュをメモリに保存しました: 作成時刻 %v", data.CreatedAt)
}

// MakeCache はGraphQLクライアントを使用してデータを取得し、キャッシュデータを生成します。
func MakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, map[string]graphql.ID, error) {
	projectNumber, err := strconv.Atoi(os.Getenv("PROJECTV2_NUMBER"))
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("PROJECTV2_NUMBERをintに変換できません: %v", err)
	}
	var baseInfo query.GetOrganizationProjectBaseInfoQuery
	// キャッシュがない場合、GraphQLクエリを実行してデータを取得
	err = client.Query(context.Background(), &baseInfo, map[string]interface{}{
		"projectNumber": graphql.Int(projectNumber),
		"organization":  graphql.String(os.Getenv("ORGANIZATION_NAME")),
	})
	if err != nil {
		return "", nil, nil, nil, err
	}

	projectId := baseInfo.Organization.ProjectV2.Id

	fieldsDict := make(map[string]graphql.ID)
	statusOptionsDict := make(map[string]graphql.ID)
	for _, field := range baseInfo.Organization.ProjectV2.Fields.Nodes {
		fieldsDict[field.ProjectV2Field.Name] = graphql.ID(field.ProjectV2Field.Id)
	}
	for _, field := range baseInfo.Organization.ProjectV2.Fields.Nodes {
		log.Println(field.ProjectV2SingleSelectField.Options, field.ProjectV2Field.Name)
		if field.ProjectV2Field.Name == "Status" {
			for _, option := range field.ProjectV2SingleSelectField.Options {
				statusOptionsDict[option.Name] = graphql.ID(option.Id)
			}
		}
	}

	// プロジェクトに紐づけられたIssue情報の取得
	var itemsInfo query.GetOrganizationProjectItemsQuery
	issuesDict := make(map[int]string)
	pageSize := 50
	var cursor string
	totalItems := 0

	for {
		log.Printf("Fetching issues page (size: %d, after: %s)", pageSize, cursor)
		err = client.Query(context.Background(), &itemsInfo, map[string]interface{}{
			"projectNumber": graphql.Int(projectNumber),
			"organization":  graphql.String(os.Getenv("ORGANIZATION_NAME")),
			"first":         graphql.Int(pageSize),
			"after":         graphql.String(cursor),
		})
		if err != nil {
			log.Printf("Issuesの取得に失敗しました: %v", err)
			return "", nil, nil, nil, err
		}

		for _, item := range itemsInfo.Organization.ProjectV2.Items.Nodes {
			issuesDict[item.Content.Issue.Number] = item.Id
			totalItems++
		}

		if !itemsInfo.Organization.ProjectV2.Items.PageInfo.HasNextPage {
			break
		}
		cursor = itemsInfo.Organization.ProjectV2.Items.PageInfo.EndCursor
		log.Printf("これまでに取得したIssue数: %d", totalItems)
	}
	log.Printf("issuesDict: %v", issuesDict)
	log.Printf("全Issueの取得が完了しました (合計: %d)", totalItems)

	return projectId, issuesDict, fieldsDict, statusOptionsDict, nil
}

// LoadOrMakeCache はキャッシュを読み込み、存在しないか期限切れの場合は新たにキャッシュを生成します。
func LoadOrMakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, map[string]graphql.ID, error) {
	baseInfo, err := LoadCache()
	if err == nil {
		return baseInfo.ID, baseInfo.IssuesDict, baseInfo.FieldsDict, baseInfo.StatusOptionsDict, nil
	}

	projectId, issuesDict, fieldsDict, statusOptionsDict, err := MakeCache(client)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("キャッシュの生成に失敗しました: %v", err)
	}
	newCache := CacheData{
		ID:                projectId,
		IssuesDict:        issuesDict,
		FieldsDict:        fieldsDict,
		StatusOptionsDict: statusOptionsDict,
		CreatedAt:         time.Now(),
	}
	SaveCache(newCache)
	return projectId, issuesDict, fieldsDict, statusOptionsDict, nil
}
