package cache

import (
	"context"
	"encoding/json"
	"os"

	graphql "github.com/mathsuky/BOT_remind/query"
)

const cacheFilePath = "cache.json"

type CacheData struct {
	ID         string
	IssuesDict map[int]string
	FieldsDict map[string]graphql.ID
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

func SaveCache(id string, dic1 map[int]string, dic2 map[string]graphql.ID) error {
	cacheData := CacheData{
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

func MakeCache(client *graphql.Client) (string, map[int]string, map[string]graphql.ID, error) {
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
