package database

import (
	"time"

	"gorm.io/gorm/logger"
)

// Options holds optional configuration for database initialization
type Options struct {
	// LogLevel overrides the default GORM log level (Warn)
	LogLevel logger.LogLevel

	// SlowThreshold overrides the default slow query threshold (200ms)
	SlowThreshold time.Duration

	// MaxOpenConns overrides the default maximum open connections (25)
	MaxOpenConns int

	// MaxIdleConns overrides the default maximum idle connections (5)
	MaxIdleConns int

	// ConnMaxLifetime overrides the default connection lifetime (30m)
	ConnMaxLifetime time.Duration
}

// DefaultOptions returns the default configuration options
func DefaultOptions() *Options {
	return &Options{
		LogLevel:        logger.Warn,
		SlowThreshold:   200 * time.Millisecond,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
	}
}

// TODO: Implement InitWithOptions(ctx, cfg, opts) when customization needed
// This would allow overriding the hardcoded values in Init() function
