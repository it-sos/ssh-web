package web

import "embed"

//go:embed index.html totp.html terminal.html css/* js/* vendor/*
var StaticFiles embed.FS
