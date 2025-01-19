package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
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

		if update.Message.Text == "/start" || update.Message.Text == "/help" {
			text := "Добро пожаловать!  Команды:  - /add user <username>  - /link user <username>  - /del user <username>  - /users"
			sendMessage(bot, update.Message.Chat.ID, text)
			continue
		} else if strings.HasPrefix(update.Message.Text, "/add user ") {
			// Извлекаем username из команды
			username := strings.TrimPrefix(update.Message.Text, "/add user ")

			command := "echo Adding Xray user: " + username + " && cd /root/xray && bash ex.sh add " + username + " && echo Xray user " + username + " added"
			// Выполняем команду с переданным username
			sendCommand(bot, update.Message.Chat.ID, command)

			command = "cd /root/xray && bash ex.sh  link conf/config_client_" + username + ".json"
			output := sendCommand(bot, update.Message.Chat.ID, command)

			// Ищем индекс начала строки "vless://"
			startIndex := strings.Index(string(output), "vless://")
			if startIndex == -1 {
				fmt.Println("Prefix 'vless://' not found.")
				continue
			}
			// Извлекаем строку начиная с "vless://"
			textMessage := "Ссылка пользователя:  `" + string(output)[startIndex:] + "`"

			sendMessage(bot, update.Message.Chat.ID, textMessage)

		} else if strings.HasPrefix(update.Message.Text, "/link user ") {
			// Извлекаем username из команды
			username := strings.TrimPrefix(update.Message.Text, "/link user ")

			command := "cd /root/xray && bash ex.sh  link conf/config_client_" + username + ".json"
			output := sendCommand(bot, update.Message.Chat.ID, command)

			// Ищем индекс начала строки "vless://"
			startIndex := strings.Index(string(output), "vless://")
			if startIndex == -1 {
				fmt.Println("Prefix 'vless://' not found.")
				continue
			}
			// Извлекаем строку начиная с "vless://"
			textMessage := "Ссылка пользователя:  `" + string(output)[startIndex:] + "`"

			sendMessage(bot, update.Message.Chat.ID, textMessage)
		} else if strings.HasPrefix(update.Message.Text, "/del user ") {
			// Извлекаем username из команды
			username := strings.TrimPrefix(update.Message.Text, "/del user ")

			command := "echo Deleting Xray user: " + username + " && cd /root/xray && bash ex.sh del " + username + " && echo Xray user " + username + " deleted"
			output := sendCommand(bot, update.Message.Chat.ID, command)
			if output != nil {
				sendMessage(bot, update.Message.Chat.ID, string(output))
			}
		} else if update.Message.Text == "/users" {
			command := "ls /root/xray/conf"
			output := sendCommand(bot, update.Message.Chat.ID, command)
			// Регулярное выражение для поиска нужных частей
			re := regexp.MustCompile(`config_client_([a-zA-Z0-9_]+)_.*\.json`)

			// Находим все совпадения
			matches := re.FindAllStringSubmatch(string(output), -1)

			// Создаем список в формате Markdown
			var result []string
			for _, match := range matches {
				// match[1] — это захваченная часть, которая содержит нужное имя
				result = append(result, "- "+match[1])
			}

			sendMessage(bot, update.Message.Chat.ID, strings.Join(result, "\n"))
		} else {
			output := sendCommand(bot, update.Message.Chat.ID, update.Message.Text)
			if output != nil {
				sendMessage(bot, update.Message.Chat.ID, string(output))
			}
		}
	}
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func sendCommand(bot *tgbotapi.BotAPI, chatID int64, command string) []byte {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Если произошла ошибка при выполнении команды, отправляем ее обратно
		msg := tgbotapi.NewMessage(chatID, "Error: "+err.Error())
		_, err := bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
		return nil
	}

	return output
}
