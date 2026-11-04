package models

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
)

var ErrReservedUsername = errors.New("Username is reserved; choose another name")

// IsReservedUsername applies only to new account creation. Local maintainer
// names admin/administrator remain available; existing accounts are not renamed.
func IsReservedUsername(username string) bool {
	switch strings.ToLower(strings.TrimSpace(username)) {
	case "api", "contact", "fasten", "chartrove", "help", "info", "login", "mail",
		"noreply", "postmaster", "root", "security", "support", "system", "webmaster", "www":
		return true
	default:
		return false
	}
}

type User struct {
	ModelBase
	FullName string `json:"full_name"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`

	//additional optional metadata that Fasten stores with users
	Picture string       `json:"picture"`
	Email   string       `json:"email"`
	Role    pkg.UserRole `json:"role"`
}

func (user *User) HashPassword(password string) error {
	if len(strings.TrimSpace(password)) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}
func (user *User) CheckPassword(providedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}
