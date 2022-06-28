package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	BotToken = os.Getenv("BOT_TOKEN")
)

func main() {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	var count int
	startTime := time.Now()

	for update := range bot.GetUpdatesChan(u) {
		if update.Message != nil {
			if update.Message.Text == "/c" {
				msg := tgbotapi.NewMessage(
					update.Message.Chat.ID,
					fmt.Sprintf("Uptime: %s\nRequests: %d", time.Since(startTime), count),
				)
				_, err = bot.Send(msg)
				if err != nil {
					log.Println(err)
				}
				continue
			}
			count += 1
			text := "Бот остановлен на неопределенный срок, по причине нехватки ресурсов для работы всех наших проектов.\n\n" +
				"Все ресурсы задействованы в нашем боте для улучшения фотографий, попробуйте бесплатно:\n@deeppaintbot \n" +
				"За новостями следите тут: @deepfaker\nСписок других наших ботов и ресурсов: t.me/deepfaker/2128"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			msg.DisableWebPagePreview = true
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
