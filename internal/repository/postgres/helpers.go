package postgres

import (
	"strings"
)

// isUniqueViolation returns true for PostgreSQL error code 23505 (unique_violation).
// We inspect the error string because pq and pgx surface errors differently.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "duplicate key")
}
