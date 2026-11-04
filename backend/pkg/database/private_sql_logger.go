package database

import (
	"context"
	"gorm.io/gorm/logger"
)

// Retain SQL structure/timing/row diagnostics without sensitive parameter values.
// Use ordinary query/create callbacks; GORM Scan's recorder interpolates early.
type privateSQLLogger struct{ logger.Interface }

func (privateSQLLogger) ParamsFilter(_ context.Context, sql string, _ ...interface{}) (string, []interface{}) {
	return sql, nil
}
func (l privateSQLLogger) LogMode(level logger.LogLevel) logger.Interface {
	return privateSQLLogger{l.Interface.LogMode(level)}
}
