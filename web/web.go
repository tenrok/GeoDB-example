package web

import "embed"

//go:embed all:static templates/*.gohtml
var FS embed.FS
