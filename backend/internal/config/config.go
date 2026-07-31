package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	JWTSecret    string
	JWTExpiresIn int
	ServerPort   string
}

func Load() *Config {
	// 从 .env 加载配置（不存在则忽略）
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "3306"),
		DBUser:       getEnv("DB_USER", "root"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       getEnv("DB_NAME", "softteststudyt"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiresIn: getEnvInt("JWT_EXPIRES_IN", 168),
		ServerPort:   getEnv("SERVER_PORT", "8080"),
	}

	if cfg.DBPassword == "" {
		log.Println("警告: 未配置 DB_PASSWORD 环境变量，数据库连接将失败")
	}
	if cfg.JWTSecret == "dev-secret-change-me" {
		log.Println("警告: 使用默认 JWT 密钥，生产环境请通过 JWT_SECRET 环境变量配置强密钥")
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
