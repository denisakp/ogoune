// Package repository provides persistence layer abstractions and implementations.
//
// This package defines repository interfaces for all domain aggregates and provides
// concrete implementations using GORM with PostgreSQL.
//
// # Error Semantics
//
// All repository methods return standard Go errors. Common repository errors are:
//   - ErrNotFound: record not found in storage
//   - ErrDuplicate: duplicate key or unique constraint violation
//   - ErrInvalidInput: invalid input data or validation failure
//
// Errors are wrapped with context using fmt.Errorf("context: %w", err) at boundaries.
//
// # Pagination
//
// All List operations support limit/offset pagination. The caller is responsible
// for providing validated pagination parameters. Repositories assume:
//   - limit > 0 and <= maximum allowed (typically 200)
//   - offset >= 0
//   - ordering is consistent (usually by created_at DESC)
//
// # Soft Delete
//
// ResourceRepository uses soft delete semantics (active=false) to preserve
// historical data and foreign key relationships. Other repositories use hard
// delete unless retention policies require otherwise.
//
// # Transaction Support
//
// Repository implementations accept a *gorm.DB instance which can be a transaction.
// This allows callers to manage transaction boundaries at the service layer.
package repository
