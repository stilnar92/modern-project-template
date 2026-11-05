// Package migrations holds the SQL migrations of the core (goose format).
// They are built into the migrate binary, so the image needs no extra files.
package migrations

import "embed"

// FS holds all migration files.
//
//go:embed *.sql
var FS embed.FS
