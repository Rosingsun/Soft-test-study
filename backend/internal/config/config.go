package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	JWTSecret            string
	JWTExpiresIn         int
	ServerPort           string
	SMTPHost             string
	SMTPPort             int
	SMTPUser             string
	SMTPPassword         string
	SMTPFromName         string
	AdminBypassUsernames []string
}

func Load() *Config {
	// 从 .env 加载配置（不存在则忽略）
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBUser:               getEnv("DB_USER", "root"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               getEnv("DB_NAME", "softteststudyt"),
		JWTSecret:            getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiresIn:         getEnvInt("JWT_EXPIRES_IN", 168),
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnvInt("SMTP_PORT", 465),
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPassword:         os.Getenv("SMTP_PASSWORD"),
		SMTPFromName:         getEnv("SMTP_FROM_NAME", "软考学系"),
		AdminBypassUsernames: getEnvCSV("ADMIN_BYPASS_USERNAMES", []string{"ross"}),
	}

	if cfg.DBPassword == "" {
		log.Println("警告: 未配置 DB_PASSWORD 环境变量，数据库连接将失败")
	}
	if cfg.JWTSecret == "dev-secret-change-me" {
		log.Println("警告: 使用默认 JWT 密钥，生产环境请通过 JWT_SECRET 环境变量配置强密钥")
	}
	if len(cfg.AdminBypassUsernames) > 0 {
		log.Printf("[config] 管理员绕过用户名: %v", cfg.AdminBypassUsernames)
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

// getEnvCSV 解析逗号分隔的字符串为 []string，自动 trim 空白与空值
func getEnvCSV(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
