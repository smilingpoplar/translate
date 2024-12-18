package config

import "embed"

//go:embed services.yaml prompt.txt
var embedFS embed.FS
