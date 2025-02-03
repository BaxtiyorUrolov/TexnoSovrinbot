package storage

import (
	"database/sql"
	"errors"
	"log"
	"texnosovrinbot/models"
	"time"

	_ "github.com/lib/pq"
)

func OpenDatabase(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetBotToken(db *sql.DB) string {
	query := `SELECT bot_token FROM configs`
	var botToken string
	err := db.QueryRow(query).Scan(&botToken)
	if err != nil {
		return ""
	}
	return botToken
}

func AddUserToDatabase(db *sql.DB, userID int) error {
	query := `INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`
	_, err := db.Exec(query, userID)
	return err
}

func AddChannelToDatabase(db *sql.DB, channelLink string) error {
	query := `INSERT INTO channels (name) VALUES ($1)`
	_, err := db.Exec(query, channelLink)
	return err
}

func GetChannelsFromDatabase(db *sql.DB) ([]string, error) {
	query := `SELECT name FROM channels`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, nil
}

func AddAdminToDatabase(db *sql.DB, adminID int64) error {
	query := `INSERT INTO admins (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`
	_, err := db.Exec(query, adminID)
	return err
}

func RemoveAdminFromDatabase(db *sql.DB, adminID int64) error {
	query := `DELETE FROM admins WHERE id = $1`
	_, err := db.Exec(query, adminID)
	return err
}

func IsAdmin(userID int, db *sql.DB) bool {
	var id int
	query := `SELECT id FROM admins WHERE id = $1`
	err := db.QueryRow(query, userID).Scan(&id)
	return err == nil
}

func GetAdmins(db *sql.DB) ([]int64, error) {
	query := `SELECT id FROM admins`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adminIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error scanning admin ID: %v", err)
			continue
		}
		adminIDs = append(adminIDs, id)
	}

	return adminIDs, nil
}

func DeleteChannelFromDatabase(db *sql.DB, channel string) error {
	query := `DELETE FROM channels WHERE name = $1`
	_, err := db.Exec(query, channel)
	return err
}

func GetTotalUsers(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

func GetTodayUsers(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= $1", time.Now().Truncate(24*time.Hour)).Scan(&count)
	return count, err
}

func GetLastMonthUsers(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= $1", time.Now().AddDate(0, -1, 0)).Scan(&count)
	return count, err
}

func GetAllUsers(db *sql.DB) ([]models.User, error) {
	log.Println("GetAllUsers funksiyasi ishga tushdi") // Log qo'shish
	query := `SELECT user_id FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func GetBalance(db *sql.DB, userID int) int {
	query := `SELECT balance FROM users WHERE id = $1`
	var balance int
	err := db.QueryRow(query, userID).Scan(&balance)
	if err != nil {
		return 0
	}
	return balance
}

func GetPrice(db *sql.DB) int {
	query := `SELECT price FROM configs LIMIT 1`
	var price int
	err := db.QueryRow(query).Scan(&price)
	if err != nil {

		return 0
	}

	return price
}

func UpdateUserInfo(db *sql.DB, telegramID int64, name, viloyat, shahar string) error {
	query := `UPDATE users SET name = $1, viloyat = $2, shahar = $3, telefon = $4 WHERE id = $5`
	result, err := db.Exec(query, name, viloyat, shahar, telegramID)
	if err != nil {
		log.Printf("Error updating user info: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected() // ✅ Nechta qator o‘zgarganini tekshirish
	if err != nil {
		log.Printf("Error getting affected rows: %v", err)
		return err
	}

	if rowsAffected == 0 { // ✅ Agar hech qanday qator o‘zgarmagan bo‘lsa, foydalanuvchi yo‘q
		log.Printf("User with id %d not found", telegramID)
		return errors.New("user not found") // ❌ Xato qaytarish
	}

	return nil
}
