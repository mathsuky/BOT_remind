package cache

import (
	"encoding/json"
	"os"
	"github.com/mathsuky/BOT_remind/graphql"
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
