package main

import (
	"context"
	"log"
	"log/slog"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Enabled(context.Background(), slog.LevelDebug)

	discordToken := os.Getenv("AVINA_DISCORD_TOKEN")
	if discordToken == "" {
		log.Fatal("Discord token not set")
	}
	discordAppId := os.Getenv("AVINA_DISCORD_APP_ID")
	if discordAppId == "" {
		log.Fatal("Discord app id not set")
	}

	avinaBot := NewBot(logger, discordToken, discordAppId)
	err := avinaBot.Start()
	if err != nil {
		log.Fatal(err)
	}
}
