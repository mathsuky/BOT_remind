package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mathsuky/BOT_remind/bot"
	"github.com/mathsuky/BOT_remind/github"
)



func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file")
	}

	client, err := github.GetClient(os.Getenv("GITHUB_TOKEN_CLASSIC"))
	if err != nil {
		log.Fatalf("failed to get client: %v", err)
	}

	accessToken := os.Getenv("TRAQ_BOT_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatalf("TRAQ_BOT_ACCESS_TOKEN is not set")
	}

	if err := bot.Start(client, accessToken); err != nil {
		panic(err)
	}
}
