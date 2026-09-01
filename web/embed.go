package web

import "embed"

// Dist 内嵌前端构建产物
//
//go:embed all:dist
var Dist embed.FS
