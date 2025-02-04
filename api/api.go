package api

import (
	"database/sql"
	"net/http"
	"texnosovrinbot/storage"

	"github.com/gin-gonic/gin"
)

type UserUpdateRequest struct {
	TelegramID int64  `json:"telegram_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Viloyat    string `json:"viloyat" binding:"required"`
	Shahar     string `json:"shahar" binding:"required"`
	Telefon    string `json:"telefon" binding:"required"`
}

// @Summary Update user info
// @Description Foydalanuvchi ma’lumotlarini yangilash
// @Tags users
// @Accept json
// @Produce json
// @Param user body UserUpdateRequest true "User data"
// @Success 200 {object} map[string]string "message: User updated successfully"
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 500 {object} map[string]string "error: Database error"
// @Router /update_user [post]
func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserUpdateRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		err := storage.UpdateUserInfo(db, req.TelegramID, req.Name, req.Viloyat, req.Shahar, req.Telefon)
		if err != nil {
			if err.Error() == "user not found" { // 🔹 `RowsAffected() == 0` bo‘lsa
				c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user info"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
	}
}
