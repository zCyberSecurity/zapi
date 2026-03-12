package config

import "os"

type Config struct {
	DBPath     string
	Addr       string
	AdminToken string
}

func Load() *Config {
	return &Config{
		DBPath:     getEnv("DB_PATH", "zapi.db"),
		Addr:       getEnv("ADDR", ":8080"),
		AdminToken: getEnv("ADMIN_TOKEN", "change-me"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
