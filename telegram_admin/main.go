package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные из файла .env
	if err := godotenv.Load(); err != nil {
		log.Panic("Error loading .env file")
	}

	// Получаем токен и разрешенный ID пользователя из .env
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Panic("TELEGRAM_BOT_TOKEN is not set in .env file")
	}

	allowedUserIDStr := os.Getenv("ALLOWED_TELEGRAM_USER_ID")
	if allowedUserIDStr == "" {
		log.Panic("ALLOWED_TELEGRAM_USER_ID is not set in .env file")
	}

	allowedUserID, err := strconv.Atoi(allowedUserIDStr)
	if err != nil {
		log.Panic("Invalid ALLOWED_TELEGRAM_USER_ID in .env file: must be an integer")
	}

	// Создаем бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Получаем обновления от Telegram
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil { // игнорируем любые обновления, которые не являются сообщениями
			continue
		}

		// Проверяем, что сообщение отправлено разрешенным пользователем
		if update.Message.From.ID != int64(allowedUserID) {
			log.Printf("Unauthorized user: %d", update.Message.From.ID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Access denied.")
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
			continue
		}

		log.Printf("[Received] From: %d, Command: %s", update.Message.From.ID, update.Message.Text)

		if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the bot!")
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
			continue
		}

		// Проверяем, что команда начинается с /add user
		if strings.HasPrefix(update.Message.Text, "/add user ") {
			// Извлекаем username из команды
			username := strings.TrimPrefix(update.Message.Text, "/add user ")

			// Выполняем команду с переданным username
			cmd := exec.Command("bash", "-c", "echo Adding Xray user: "+username+" && cd /root/xray && bash ex.sh add "+username+" && echo Xray user "+username+" added") // Здесь замените на вашу команду
			output, err := cmd.CombinedOutput()
			if err != nil {
				// Если произошла ошибка при выполнении команды, отправляем ее обратно
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
				_, err := bot.Send(msg)
				if err != nil {
					log.Printf("Error sending message: %v", err)
				}
				continue
			}

			cmd = exec.Command("bash", "-c", "cd /root/xray && bash ex.sh  link conf/config_client_"+username+".json") // Здесь замените на вашу команду
			output, err = cmd.CombinedOutput()
			if err != nil {
				// Если произошла ошибка при выполнении команды, отправляем ее обратно
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
				_, err := bot.Send(msg)
				if err != nil {
					log.Printf("Error sending message: %v", err)
				}
				continue
			}

			// Ищем индекс начала строки "vless://"
			startIndex := strings.Index(string(output), "vless://")
			if startIndex == -1 {
				fmt.Println("Prefix 'vless://' not found.")
				continue
			}

			// Извлекаем строку начиная с "vless://"
			textMessage := string(output)[startIndex:]
			// Отправляем результат выполнения команды обратно в Telegram
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ссылка пользователя: `"+textMessage+"`")
			msg.ParseMode = "markdown"
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}

		if strings.HasPrefix(update.Message.Text, "/link user ") {
			// Извлекаем username из команды
			username := strings.TrimPrefix(update.Message.Text, "/link user ")

			cmd := exec.Command("bash", "-c", "cd /root/xray && bash ex.sh  link conf/config_client_"+username+".json") // Здесь замените на вашу команду
			output, err := cmd.CombinedOutput()
			if err != nil {
				// Если произошла ошибка при выполнении команды, отправляем ее обратно
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
				_, err := bot.Send(msg)
				if err != nil {
					log.Printf("Error sending message: %v", err)
				}
				continue
			}

			// Ищем индекс начала строки "vless://"
			startIndex := strings.Index(string(output), "vless://")
			if startIndex == -1 {
				fmt.Println("Prefix 'vless://' not found.")
				continue
			}

			// Извлекаем строку начиная с "vless://"
			textMessage := string(output)[startIndex:]
			// Отправляем результат выполнения команды обратно в Telegram
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ссылка пользователя: `"+textMessage+"`")
			msg.ParseMode = "markdown"
			_, err = bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}

		// Выполняем команду из текста сообщения
		cmd := exec.Command("bash", "-c", update.Message.Text)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Если произошла ошибка при выполнении команды, отправляем ее обратно
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error: "+err.Error())
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}
			continue
		}

		// Отправляем результат выполнения команды обратно в Telegram
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, string(output))
		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
