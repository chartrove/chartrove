package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fastenhealth/fasten-onprem/backend/pkg"
	"github.com/fastenhealth/fasten-onprem/backend/pkg/models"
	"gorm.io/gorm"
)

var ErrRegistrationClosed = errors.New("public registration is disabled")

// RegisterUser serializes the count/policy/role decision and all account
// initialization writes across connections/processes. CreateUser remains the
// independent administrator-managed creation path, with its existing contract.
func (gr *GormRepository) RegisterUser(ctx context.Context, user *models.User, enabled bool) error {
	if user == nil || strings.TrimSpace(user.Username) == "" || strings.TrimSpace(user.Password) == "" {
		return errors.New("username and password are required")
	}
	if models.IsReservedUsername(user.Username) {
		return models.ErrReservedUsername
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return err
	}
	if !enabled {
		// This read only avoids password work for an already-closed policy.
		// Bootstrap races are decided again inside the serialized transaction.
		count, err := gr.GetUserCount(ctx)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrRegistrationClosed
		}
	}
	if err := user.HashPassword(user.Password); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	options := &sql.TxOptions{}
	if gr.GormClient.Dialector.Name() == "postgres" {
		// The count after a waited advisory lock needs a fresh statement
		// snapshot even if the operator's default is REPEATABLE READ.
		options.Isolation = sql.LevelReadCommitted
	}
	return gr.GormClient.WithContext(ctx).Session(&gorm.Session{Logger: privateSQLLogger{gr.GormClient.Logger}}).Transaction(func(tx *gorm.DB) error {
		switch tx.Dialector.Name() {
		case "sqlite":
			// Take SQLite's writer reservation before the first read, even for an
			// empty table. This changes no user row or timestamp.
			if err := tx.Exec("UPDATE users SET id = id WHERE 1 = 0").Error; err != nil {
				return err
			}
		case "postgres":
			// Application-specific transaction lock; no table, persistent lock,
			// privilege change or production infrastructure is introduced.
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(0x4368727472765247)).Error; err != nil {
				return err
			}
		default:
			return fmt.Errorf("registration unavailable for this database")
		}
		var count int64
		if err := tx.Model(&models.User{}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 && !enabled {
			return ErrRegistrationClosed
		}
		user.Role = pkg.UserRoleUser
		if count == 0 {
			user.Role = pkg.UserRoleAdmin
		}
		repo := *gr
		repo.GormClient = tx
		return repo.createUserRecords(ctx, user)
	}, options)
}
