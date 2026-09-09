// Package migrations embeds the numbered SQL schema migrations so slamctl is a
// self-contained binary — the .sql files ship inside it and cannot drift from
// the code that runs them.
//
// File naming (golang-migrate): {version}_{title}.up.sql plus a paired
// .down.sql. Versions are monotonic and a committed migration is NEVER edited —
// a change is always a new, higher-numbered pair.
package migrations

import "embed"

// FS holds every .sql migration in this directory.
//
//go:embed *.sql
var FS embed.FS
