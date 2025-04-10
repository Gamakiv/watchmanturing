package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структура для хранения настроек из config.json
type Config struct {
	BotToken          string `json:"bot_token"`            // Токен Telegram-бота
	CheckCursorMillis string `json:"check_cursor_setting"` // Интервал проверки курсора (мс)
}

// Глобальные переменные для управления режимом охраны
var cursorWatcher *CursorWatcher
var watcherMutex sync.Mutex

// Загрузка конфигурации из файла
func loadConfig(filename string) (*Config, error) {
	// Проверяем существование файла
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл конфигурации %s не найден", filename)
	}

	// Открываем файл
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла конфигурации: %v", err)
	}
	defer file.Close()

	// Декодируем JSON в структуру Config
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла конфигурации: %v", err)
	}

	// Проверяем валидность данных
	if config.BotToken == "" {
		return nil, fmt.Errorf("токен бота не задан в конфигурации")
	}

	return &config, nil
}

// Функция для обработки перемещения курсора
func onCursorMoved(message string) {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	if bot != nil && chatID != 0 {
		msg := tgbotapi.NewMessage(chatID, message)
		bot.Send(msg)
	}
}

// Глобальные переменные для Telegram-бота
var bot *tgbotapi.BotAPI
var chatID int64

func main() {
	// Загружаем конфигурацию
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Ошибка при загрузке конфигурации: %v", err)
	}

	// Преобразуем интервал проверки курсора в число
	checkMillis, err := strconv.Atoi(config.CheckCursorMillis)
	if err != nil || checkMillis <= 0 {
		log.Fatalf("Неверное значение check_cursor_setting в конфигурации: %v", err)
	}

	// Создание экземпляра бота
	botInstance, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}
	bot = botInstance

	// Вывод информации о боте
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Создание клавиатуры с кнопками меню
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Проверка"),
			tgbotapi.NewKeyboardButton("Блокировка"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Режим охраны"),
		),
	)

	// Настройка обновлений (polling)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	//updates, err := bot.GetUpdatesChan(u)
	updates := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Ошибка при получении обновлений: %v", err)
	}

	// Обработка входящих сообщений
	for update := range updates {
		if update.Message == nil { // Игнорируем несообщения
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// Сохраняем ID чата для отправки уведомлений
		chatID = update.Message.Chat.ID

		// Отправляем клавиатуру при первом сообщении
		if update.Message.IsCommand() && update.Message.Command() == "start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать! Выберите действие:")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)
			continue
		}

		// Обработка нажатий на кнопки
		switch update.Message.Text {
		case "Проверка":
			handleCheck(bot, update.Message)
		case "Блокировка":
			handleLock(bot, update.Message)
		case "Режим охраны":
			handleGuardMode(bot, update.Message, checkMillis)
		default:
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Пожалуйста, используйте кнопки.")
			bot.Send(msg)
		}
	}
}

// Функция для обработки кнопки "Проверка"
func handleCheck(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Выполняю проверку состояния экрана...")
	bot.Send(msg)

	// Проверяем состояние экрана
	screenStatus := getScreenStatus()

	// Отправляем результат
	msg.Text = screenStatus
	bot.Send(msg)
}

// Функция для обработки кнопки "Блокировка"
func handleLock(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	// Проверяем, заблокирован ли уже экран
	if isScreenLocked() {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Экран уже заблокирован.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Выполняю блокировку системы...")
	bot.Send(msg)

	// Блокируем экран
	err := lockScreen()
	if err != nil {
		msg.Text = "Ошибка при блокировке экрана: " + err.Error()
	} else {
		msg.Text = "Система заблокирована."
	}

	bot.Send(msg)
}

// Функция для обработки кнопки "Режим охраны"
func handleGuardMode(bot *tgbotapi.BotAPI, message *tgbotapi.Message, checkMillis int) {
	watcherMutex.Lock()
	defer watcherMutex.Unlock()

	if cursorWatcher == nil {
		cursorWatcher = NewCursorWatcher(checkMillis, 500, 300)
	}

	if !cursorWatcher.isRunning {
		cursorWatcher.Start()
		msg := tgbotapi.NewMessage(message.Chat.ID, "Режим охраны активирован. Отслеживаю курсор...")
		bot.Send(msg)
	} else {
		cursorWatcher.Stop()
		msg := tgbotapi.NewMessage(message.Chat.ID, "Режим охраны деактивирован.")
		bot.Send(msg)
	}
}
