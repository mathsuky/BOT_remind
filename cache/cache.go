package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/query"
)

const cacheFilePath = "cache.json"

type CacheData struct {
	ID                string
	IssuesDict        map[int]string
	FieldsDict        map[string]graphql.ID
	StatusOptionsDict map[string]graphql.ID
}

func LoadCache() (CacheData, error) {
	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return CacheData{}, err
	}

	var cache CacheData
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return CacheData{}, err
	}

	return cache, nil
}

func SaveCache(id string, dic1 map[int]string, dic2 map[string]graphql.ID, dic3 map[string]graphql.ID) error {
	cacheData := CacheData{
		ID:                id,
		IssuesDict:        dic1,
		FieldsDict:        dic2,
		StatusOptionsDict: dic3,
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

func MakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, map[string]graphql.ID, error) {
	projectNumber, err := strconv.Atoi(os.Getenv("PROJECTV2_NUMBER"))
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("failed to convert PROJECTV2_NUMBER to int: %v", err)
	}
	var info query.GetProjectBaseInfoQuery
	// キャッシュがない場合はクエリを実行してキャッシュを保存
	err = client.Query(context.Background(), &info, map[string]interface{}{
		"projectNumber": graphql.Int(projectNumber),
		"user":          graphql.String(os.Getenv("REPOSITORY_OWNER")),
	})
	if err != nil {
		return "", nil, nil, nil, err
	}
	log.Println(info.User.ProjectV2.Items.Nodes)

	projectId := info.User.ProjectV2.Id
	issuesDict := make(map[int]string)
	for _, item := range info.User.ProjectV2.Items.Nodes {
		issuesDict[item.Content.Issue.Number] = item.Id
	}
	fieldsDict := make(map[string]graphql.ID)
	for _, field := range info.User.ProjectV2.Fields.Nodes {
		fieldsDict[field.ProjectV2Field.Name] = graphql.ID(field.ProjectV2Field.Id)
	}
	statusOptionsDict := make(map[string]graphql.ID)
	for _, field := range info.User.ProjectV2.Fields.Nodes {
		log.Println(field.ProjectV2SingleSelectField.Options, field.ProjectV2Field.Name)
		if field.ProjectV2Field.Name == "Status" {
			for _, option := range field.ProjectV2SingleSelectField.Options {
				statusOptionsDict[option.Name] = graphql.ID(option.Id)
			}
		}
	}

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
