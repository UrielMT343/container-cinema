package models

type User struct {
	Id           int    `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	PasswordHash string `db:"password_hash" json:"passwrdHash"`
	IsActive     bool   `db:"is_active" json:"isActive"`
}
