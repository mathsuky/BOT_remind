package github

import (
	"fmt"
	"net/http"

	"github.com/hasura/go-graphql-client"
	"github.com/mathsuky/BOT_remind/transport"
)

func GetClient(token string) (*graphql.Client, error) {
	if token == "" {
		return nil, fmt.Errorf("GitHub token is not set")
	}
	httpClient := &http.Client{
		Transport: &transport.Transport{Token: token},
	}
	return graphql.NewClient("https://api.github.com/graphql", httpClient), nil
}
