package models

import (
	"errors"
	"strings"
)

type User struct {
	ID           int    `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	Email 		 string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash" json:"passwordHash"`
	IsActive     bool   `db:"is_active" json:"isActive"`
	Role         string `db:"role" json:"role"`
}

func (u *User) Validate() error {
	var errs []string
	if u.Name == "" {
		errs = append(errs, "the name cannot be empty")
	}
	if u.Email == "" {
		errs = append(errs, "the email cannot be empty")
	}
	if u.PasswordHash == "" {
		errs = append(errs, "the pasword cannot be empty")
	}
	if u.Role == "" {
		errs = append(errs, "the role cannot be empty")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
