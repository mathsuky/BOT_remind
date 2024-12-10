package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v67/github"
	"github.com/joho/godotenv"
)

func main() {
	// .env ファイルを読み込む
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// 環境変数から統合 ID とインストール ID を取得
	appID, err := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
	if err != nil {
		log.Fatalf("Error parsing GITHUB_APP_ID: %v", err)
	}

	installationID, err := strconv.ParseInt(os.Getenv("GITHUB_INSTALLATION_ID"), 10, 64)
	if err != nil {
		log.Fatalf("Error parsing GITHUB_INSTALLATION_ID: %v", err)
	}

	privateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")

	// Wrap the shared transport for use with the integration ID and installation ID.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, installationID, privateKeyPath)
	if err != nil {
		log.Fatalf("Error creating installation transport: %v", err)
	}

	// Use installation transport with client.
	client := github.NewClient(&http.Client{Transport: itr})

	// 特定のリポジトリのイシューをリストする
	owner := "mathsuky"
	repo := "BOT_remind"
	opts := &github.IssueListByRepoOptions{
		State: "open", // open/closed/all
	}

	issues, _, err := client.Issues.ListByRepo(context.Background(), owner, repo, opts)
	if err != nil {
		log.Fatalf("Error listing issues for repository: %v", err)
	}

	// イシューを表示する
	for _, issue := range issues {
		fmt.Printf("Issue #%d: %s\n", issue.GetNumber(), issue.GetTitle())
		fmt.Printf("comments: %d\n", issue.GetComments())

		// イシューのコメントをリストする
		comments, _, err := client.Issues.ListComments(context.Background(), owner, repo, issue.GetNumber(), nil)
		if err != nil {
			log.Fatalf("Error listing comments for issue #%d: %v", issue.GetNumber(), err)
		}

		// コメントを表示する
		for _, comment := range comments {
			commentJSON, err := json.MarshalIndent(comment, "", "  ")
			if err != nil {
				log.Fatalf("Error marshalling comment to JSON: %v", err)
			}
			fmt.Println(string(commentJSON))
		}
	}
}
