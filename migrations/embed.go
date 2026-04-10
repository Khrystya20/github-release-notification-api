package migrations

import "embed"

// Files contains all SQL migrations bundled into the application binary.
//
//go:embed *.sql
var Files embed.FS
