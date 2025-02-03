package models

type User struct {
	ID      int64
	Ism     string `json:"ism"`
	Viloyat string `json:"viloyat" binding:"required"`
	Shahar  string `json:"shahar" binding:"required"`
	Telefon string `json:"telefon" binding:"required"`
}
