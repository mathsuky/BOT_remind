package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/query"
)

var cacheFilePath string

func init() {
	// キャッシュディレクトリの設定
	// 1. CACHE_DIR環境変数
	// 2. /tmp/bot-remind（デフォルト）
	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "bot-remind")
	}

	// キャッシュディレクトリの作成（0755パーミッション）
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("Warning: Failed to create cache directory: %v", err)
		// フォールバック: 一時ディレクトリ直下
		cacheDir = os.TempDir()
	}

	cacheFilePath = filepath.Join(cacheDir, "github_project.json")
	log.Printf("Using cache file: %s", cacheFilePath)
}

type CacheData struct {
	ID                string
	IssuesDict        map[int]string
	FieldsDict        map[string]graphql.ID
	StatusOptionsDict map[string]graphql.ID
	CreatedAt         time.Time
}

// キャッシュの有効期間（24時間）
const cacheTTL = 24 * time.Hour

func LoadCache() (CacheData, error) {
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		log.Printf("Failed to read cache file: %v", err)
		return CacheData{}, err
	}

	var cache CacheData
	if err = json.Unmarshal(data, &cache); err != nil {
		log.Printf("Failed to unmarshal cache data: %v", err)
		return CacheData{}, err
	}

	// キャッシュの有効期限をチェック
	if time.Since(cache.CreatedAt) > cacheTTL {
		log.Printf("Cache expired: created at %v", cache.CreatedAt)
		return CacheData{}, fmt.Errorf("cache expired")
	}

	log.Printf("Cache loaded successfully: created at %v", cache.CreatedAt)

	return cache, nil
}

func SaveCache(id string, dic1 map[int]string, dic2 map[string]graphql.ID, dic3 map[string]graphql.ID) error {
	cacheData := CacheData{
		ID:                id,
		IssuesDict:        dic1,
		FieldsDict:        dic2,
		StatusOptionsDict: dic3,
		CreatedAt:         time.Now(),
	}

	// 既存のキャッシュファイルが存在する場合は削除
	if err := os.Remove(cacheFilePath); err != nil && !os.IsNotExist(err) {
		log.Printf("Failed to remove old cache file: %v", err)
		return fmt.Errorf("failed to remove old cache file: %v", err)
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	// 0644パーミッションでファイルを作成
	err = os.WriteFile(cacheFilePath, data, 0644)
	if err != nil {
		log.Printf("Failed to write cache file: %v", err)
		return err
	}

	log.Printf("Cache saved successfully at %v", time.Now())

	return nil
}

func MakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, map[string]graphql.ID, error) {
	projectNumber, err := strconv.Atoi(os.Getenv("PROJECTV2_NUMBER"))
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("failed to convert PROJECTV2_NUMBER to int: %v", err)
	}
	var baseInfo query.GetOrganizationProjectBaseInfoQuery
	// キャッシュがない場合はクエリを実行してキャッシュを保存
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

	// projectsに紐づけられたissueの情報を取得
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
			log.Printf("Failed to fetch issues: %v", err)
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
		log.Printf("Fetched %d items so far", totalItems)
	}
	log.Printf("issuesDict: %v", issuesDict)
	log.Printf("Completed fetching all issues (total: %d)", totalItems)

	return projectId, issuesDict, fieldsDict, statusOptionsDict, nil
}

func LoadOrMakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, map[string]graphql.ID, error) {
	baseInfo, err := LoadCache()
	if err == nil {
		return baseInfo.ID, baseInfo.IssuesDict, baseInfo.FieldsDict, baseInfo.StatusOptionsDict, nil
	}

	projectId, issuesDict, fieldsDict, statusOptionsDict, err := MakeCache(client)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("failed to make cache: %v", err)
	}
	err = SaveCache(projectId, issuesDict, fieldsDict, statusOptionsDict)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("failed to save cache: %v", err)
	}
	return projectId, issuesDict, fieldsDict, statusOptionsDict, nil
}
