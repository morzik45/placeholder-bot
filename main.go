package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	var count int
	startTime := time.Now()

	for update := range updates {
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
			//text := "По причине нестабильной работы бот отправлен на переделку и вернётся в ближайшие дни.\nЗа новостями следите тут - @deepfaker\nСписок других наших ботов и ресурсов - https://t.me/DeepFaker/2128"
			text := "Бот остановлен на неопределенный срок, по причине нехватки ресурсов для работы всех наших проектов.\n\nВсе ресурсы задействованы в нашем боте для улучшения фотографий, попробуйте бесплатно:\n@deeppaintbot \nЗа новостями следите тут: @deepfaker\nСписок других наших ботов и ресурсов: t.me/deepfaker/2128"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
			msg.DisableWebPagePreview = true
			_, err = bot.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
