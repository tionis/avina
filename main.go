package main

import (
	"context"
	UniPassauBot "github.com/tionis/uni-passau-bot/api"
	"log"
	"log/slog"
	"net/http"
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

	uniPassauBotTelegramToken := os.Getenv("UNIPASSAU_BOT_TELEGRAM_TOKEN")
	if uniPassauBotTelegramToken != "" {
		go UniPassauBot.UniPassauBot(logger.WithGroup("UniPassauBot"), uniPassauBotTelegramToken)
	}

	// Start a http server on port 8080 for health checks
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		http.ListenAndServe(":8080", nil)
	}()

	avinaBot := NewBot(logger, discordToken, discordAppId)
	err := avinaBot.Start()
	if err != nil {
		log.Fatal(err)
	}
}
