package handle

import (
	"bytes"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"texnosovrinbot/admin"
	"texnosovrinbot/state"
	"texnosovrinbot/storage"
	"time"
)

func HandleUpdate(update tgbotapi.Update, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	if update.Message != nil {
		handleMessage(update.Message, db, botInstance)
	} else if update.CallbackQuery != nil {
		handleCallbackQuery(update.CallbackQuery, db, botInstance)
	} else {
		log.Printf("Unsupported update type: %T", update)
	}
}

func handleMessage(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	text := msg.Text

	log.Printf("Received message: %s", text)

	if userState, exists := state.UserStates[chatID]; exists {
		log.Printf("User state: %s", userState)
		switch userState {
		case "waiting_for_broadcast_message":
			admin.HandleBroadcastMessage(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_channel_link":
			admin.HandleChannelLink(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_admin_id":
			admin.HandleAdminAdd(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		case "waiting_for_admin_id_remove":
			admin.HandleAdminRemove(msg, db, botInstance)
			delete(state.UserStates, chatID)
			return
		}
	}

	if text == "/start" {
		handleStartCommand(msg, db, botInstance)
		storage.AddUserToDatabase(db, int(msg.Chat.ID))
	} else if text == "/admin" {
		admin.HandleAdminCommand(msg, db, botInstance)
	} else {
		handleDefaultMessage(msg, db, botInstance)
	}
}

func handleStartCommand(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	firstName := msg.From.FirstName

	log.Printf("Checking if user %d is already in database...", userID)
	err := storage.AddUserToDatabase(db, userID)
	if err != nil {
		log.Printf("Error adding user to database: %v", err)
		return
	}

	// Barcha kerakli kanallarni olish
	channels, err := storage.GetChannelsFromDatabase(db)
	if err != nil {
		log.Printf("Error getting channels from database: %v", err)
		return
	}

	// Agar foydalanuvchi barcha kanallarga azo bo'lmagan bo'lsa, kanallarga azo bo‘lishni talab qilamiz
	if !isUserSubscribedToChannels(chatID, channels, botInstance) {
		message := tgbotapi.NewMessage(chatID, "⚠️ Iltimos, quyidagi kanallarga azo bo'ling:")
		inlineKeyboard := createSubscriptionKeyboard(channels)
		message.ReplyMarkup = inlineKeyboard
		botInstance.Send(message)
		return
	}

	// Agar foydalanuvchi barcha kanallarga azo bo‘lgan bo‘lsa, unga Web App havolasini beramiz
	webAppURL := fmt.Sprintf("https://it-tanlov.uz?user_id=%d", userID)
	webAppButton := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonURL("📋 Ma'lumotlarni to‘ldirish", webAppURL),
		},
	)

	welcomeMessage := fmt.Sprintf("👋 Assalomu alaykum [%s](tg://user?id=%d), botimizga xush kelibsiz!\n\nEndi ma’lumotlaringizni to‘ldirish uchun quyidagi tugmani bosing.", firstName, userID)

	// Yangi `message` yaratamiz, lekin `:=` ishlatmaymiz
	message := tgbotapi.NewMessage(chatID, welcomeMessage)
	message.ParseMode = "Markdown"
	message.ReplyMarkup = webAppButton

	_, err = botInstance.Send(message)
	if err != nil {
		log.Printf("Error sending Web App link: %v", err)
	}
}

func handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := callbackQuery.Message.Chat.ID
	messageID := callbackQuery.Message.MessageID

	channels, err := storage.GetChannelsFromDatabase(db)
	if err != nil {
		log.Printf("Error getting channels from database: %v", err)
		return
	}

	if callbackQuery.Data == "check_subscription" {
		if isUserSubscribedToChannels(chatID, channels, botInstance) {
			deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
			botInstance.Send(deleteMsg)

			welcomeMessage := fmt.Sprintf("👋 Assalomu alaykum [%s](tg://user?id=%d) botimizga xush kelibsiz.", callbackQuery.From.FirstName, callbackQuery.From.ID)

			msg := tgbotapi.NewMessage(chatID, welcomeMessage)
			msg.ParseMode = "Markdown"
			_, err = botInstance.Send(msg)
			if err != nil {
				log.Printf("Error sending photo: %v", err)
				return
			}
		} else {
			msg := tgbotapi.NewMessage(chatID, "Iltimos, kanallarga azo bo'ling.")
			inlineKeyboard := createSubscriptionKeyboard(channels)
			msg.ReplyMarkup = inlineKeyboard
			botInstance.Send(msg)
		}
	} else if strings.HasPrefix(callbackQuery.Data, "delete_channel_") {
		channel := strings.TrimPrefix(callbackQuery.Data, "delete_channel_")
		admin.AskForChannelDeletionConfirmation(chatID, messageID, channel, botInstance)
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
		botInstance.Send(deleteMsg)
	} else if strings.HasPrefix(callbackQuery.Data, "confirm_delete_channel_") {
		channel := strings.TrimPrefix(callbackQuery.Data, "confirm_delete_channel_")
		admin.DeleteChannel(chatID, messageID, channel, db, botInstance)
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
		botInstance.Send(deleteMsg)
	} else if callbackQuery.Data == "cancel_delete_channel" {
		admin.CancelChannelDeletion(chatID, messageID, botInstance)
	}
}

func handleDefaultMessage(msg *tgbotapi.Message, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	chatID := msg.Chat.ID
	text := msg.Text

	// Check if the message is a numeric ID or a username
	if isNumeric(text) || strings.HasPrefix(text, "@") {

		return
	}

	switch text {
	case "Kanal qo'shish":
		state.UserStates[chatID] = "waiting_for_channel_link"
		msgResponse := tgbotapi.NewMessage(chatID, "Kanal linkini yuboring (masalan, https://t.me/your_channel):")
		botInstance.Send(msgResponse)
	case "Admin qo'shish":
		state.UserStates[chatID] = "waiting_for_admin_id"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, yangi admin ID sini yuboring:")
		botInstance.Send(msgResponse)
	case "Admin o'chirish":
		state.UserStates[chatID] = "waiting_for_admin_id_remove"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, admin ID sini o'chirish uchun yuboring:")
		botInstance.Send(msgResponse)
	case "Kanal o'chirish":
		admin.DisplayChannelsForDeletion(chatID, db, botInstance)
	case "Statistika":
		admin.HandleStatistics(msg, db, botInstance)
	case "Habar yuborish":
		state.UserStates[chatID] = "waiting_for_broadcast_message"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, yubormoqchi bo'lgan habaringizni kiriting (Bekor qilish uchun /cancel):")
		botInstance.Send(msgResponse)
	case "Pul tashlash":
		state.UserStates[chatID] = "waiting_for_person_id"
		msgResponse := tgbotapi.NewMessage(chatID, "Iltimos, pul tashlashingiz kerak bo'lgan odam ID sini kiriting:")
		botInstance.Send(msgResponse)
	case "BackUp olish":
		if storage.IsAdmin(int(chatID), db) {
			go HandleBackup(db, botInstance)
		}
	}
}

func isUserSubscribedToChannels(chatID int64, channels []string, botInstance *tgbotapi.BotAPI) bool {
	for _, channel := range channels {
		log.Printf("Checking subscription to channel: %s", channel)
		chat, err := botInstance.GetChat(tgbotapi.ChatConfig{SuperGroupUsername: "@" + channel})
		if err != nil {
			log.Printf("Error getting chat info for channel %s: %v", channel, err)
			return false
		}

		member, err := botInstance.GetChatMember(tgbotapi.ChatConfigWithUser{
			ChatID: chat.ID,
			UserID: int(chatID),
		})
		if err != nil {
			log.Printf("Error getting chat member info for channel %s: %v", channel, err)
			return false
		}
		if member.Status == "left" || member.Status == "kicked" {
			log.Printf("User %d is not subscribed to channel %s", chatID, channel)
			return false
		}
	}
	return true
}

func createSubscriptionKeyboard(channels []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, channel := range channels {
		channelName := strings.TrimPrefix(channel, "https://t.me/")
		button := tgbotapi.NewInlineKeyboardButtonURL("Kanalga azo bo'lish", "https://t.me/"+channelName)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	checkButton := tgbotapi.NewInlineKeyboardButtonData("Azo bo'ldim", "check_subscription")
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(checkButton))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func isNumeric(s string) bool {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}

func HandleBackup(db *sql.DB, botInstance *tgbotapi.BotAPI) {
	// Hozirgi sana
	currentTime := time.Now().Format("2006-01-02")
	backupDir := "./backups"
	backupFile := fmt.Sprintf("%s/backup_%s.sql", backupDir, currentTime)

	// Backup katalogini yaratish
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
			log.Printf("Backup katalogini yaratib bo'lmadi: %v", err)
			return
		}
	}

	// PostgreSQL backupni yaratish
	cmd := exec.Command("pg_dump", "-U", "godb", "-d", "texnosovrinbot", "-f", backupFile)
	cmd.Env = append(os.Environ(), "PGPASSWORD=0208") // Parolni muhit o'zgaruvchisi sifatida o'rnatish

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Backup yaratishda xatolik: %v, %s", err, stderr.String())
		return
	}

	log.Printf("Backup muvaffaqiyatli yaratildi: %s", backupFile)

	// Adminlarning IDlarini olish
	adminIDs, err := storage.GetAdmins(db)
	if err != nil {
		log.Printf("Adminlarni olishda xatolik: %v", err)
		return
	}

	// Faylni barcha adminlarga yuborish
	for _, chatID := range adminIDs {
		SendBackupToAdmin(chatID, backupFile, botInstance)
	}
}

// SendBackupToAdmin sends a backup file to a specific admin
func SendBackupToAdmin(chatID int64, filePath string, botInstance *tgbotapi.BotAPI) {
	// Faylni ochish
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Backup faylni ochib bo'lmadi: %v", err)
		return
	}
	defer file.Close()

	// Telegram orqali fayl yuborish
	msg := tgbotapi.NewDocumentUpload(chatID, tgbotapi.FileReader{
		Name:   filePath,
		Reader: file,
		Size:   -1,
	})

	if _, err := botInstance.Send(msg); err != nil {
		log.Printf("Admin (%d) uchun backupni yuborishda xatolik: %v", chatID, err)
	} else {
		log.Printf("Admin (%d) uchun backup muvaffaqiyatli yuborildi.", chatID)
	}
}
