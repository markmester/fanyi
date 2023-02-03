/*
 * File: main.go
 * Project: slackbot
 * File Created: Tuesday, 24th January 2023 5:26:47 pm
 * Author: Mark Mester (mmester6016@gmail.com)
 * -----
 * Last Modified: Saturday, 28th January 2023 4:53:25 pm
 * Modified By: Mark Mester (mmester6016@gmail.com>)
 */
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"

	slackbot "github.com/markmester/fanyi-slackbot/pkg/bot"
	clients "github.com/markmester/fanyi-slackbot/pkg/clients"
)

var (
	slackBotToken = getEnvOrPanic("SLACK_BOT_TOKEN")
	slackAppToken = getEnvOrPanic("SLACK_APP_TOKEN")
	chatGptApiKey = getEnvOrPanic("CHATGPT_API_KEY")
	chatGptEngine = getEnvOrPanic("CHATGPT_COMPLETION_ENGINE")

	datastorePath = os.Getenv("DATASTORE_PATH")
)

func getEnvOrPanic(env string) string {
	e := os.Getenv(env)
	if e == "" {
		panic(fmt.Sprintf("Required environmental variable unset: %s", env))
	}
	return e
}

func main() {
	// Initialize clients
	slackClient := clients.NewSlackClient(slackBotToken, slackAppToken)
	gpt3Client := clients.NewGpt3Client(chatGptApiKey, chatGptEngine)
	detector := clients.NewDetector()
	datastore, err := clients.NewDatastore(datastorePath)
	if err != nil {
		panic(err)
	}

	// Initialize bot
	bot, err := slackbot.New(slackClient, gpt3Client, detector, datastore)
	if err != nil {
		panic(err)
	}
	defer bot.Shutdown()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wake up!
	go func() {
		if err := bot.Process(); err != nil {
			panic(err)
		}
	}()

	<-sigs
}
