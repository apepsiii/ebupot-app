package main

import "embed"

//go:embed all:templates
var templateFS embed.FS

//go:embed all:public
var publicFS embed.FS

//go:embed config.yaml
var configYAML []byte
