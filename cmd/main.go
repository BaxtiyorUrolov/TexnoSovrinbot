package main

import (
	"context"
	"database/sql"
	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"texnosovrinbot/api"
	_ "texnosovrinbot/docs"
	"texnosovrinbot/handle"
	"time"

	"texnosovrinbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	shutdownTimeout = 5 * time.Second
)

func main() {
	connStr := "user=godb password=0208 dbname=texnosovrinbot sslmode=disable"
	db, err := storage.OpenDatabase(connStr)
	if err != nil {
		log.Fatal("Database connection failed: ", err)
	}
	defer db.Close()

	botToken := storage.GetBotToken(db)
	botInstance, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Failed to create Telegram bot: ", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Telegram bot gorutina
	go startTelegramBot(ctx, db, botInstance)

	// HTTP server sozlamalari
	router := gin.Default()

	// 🔹 Swagger UI'ni qo‘shish
	router.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))
	router.POST("/update_user", api.UpdateUser(db))

	server := &http.Server{
		Addr:    ":8090",
		Handler: router,
	}

	// HTTP serverni boshqarish gorutina
	go func() {
		log.Println("Starting HTTP server on :8090")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received")
}

// Start Telegram bot and listen for updates
func startTelegramBot(ctx context.Context, db *sql.DB, botInstance *tgbotapi.BotAPI) {
	offset := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping Telegram bot...")
			return
		default:
			updates, err := botInstance.GetUpdates(tgbotapi.NewUpdate(offset))
			if err != nil {
				log.Printf("Error getting updates: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for _, update := range updates {
				handle.HandleUpdate(update, db, botInstance)
				offset = update.UpdateID + 1
			}
		}
	}
}
