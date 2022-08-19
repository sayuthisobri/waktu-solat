package common

import (
	"github.com/joho/godotenv"
	"os"
)

const (
	ENV_PREFIX = "WS_"
)

type Config struct {
	IsDebug bool
	Mode    string
	DbPath  string
}
type Ctx struct {
	Config *Config
}

func (c *Ctx) LoadEnv() {
	_ = godotenv.Load()
	if os.Getenv("alfred_debug") == "1" {
		_ = os.Setenv(ENV_PREFIX+"DEBUG", "true")
	}
	if len(os.Getenv("alfred_workflow_uid")) > 0 {
		_ = os.Setenv(ENV_PREFIX+"MODE", "alfred")
	}
}

func (c *Config) IsAlfred() bool {
	return c.Mode == "alfred"
}
