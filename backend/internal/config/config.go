package config

import (
	"log"
	"os"
	"path/filepath"
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
	// OPT-01: CORS 白名单
	CORSAllowedOrigins []string
	// OPT-01: 应用环境，development / production（同时为 OPT-02 准备）
	AppEnv string
}

func Load() *Config {
	loadEnvFile()

	cfg := &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBUser:               getEnv("DB_USER", "root"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBName:               getEnv("DB_NAME", "softteststudyt"),
		JWTSecret:            getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiresIn:         getEnvInt("JWT_EXPIRES_IN", 168),
		ServerPort:           getEnv("SERVER_PORT", "3000"),
		SMTPHost:             getEnv("SMTP_HOST", ""),
		SMTPPort:             getEnvInt("SMTP_PORT", 465),
		SMTPUser:             getEnv("SMTP_USER", ""),
		SMTPPassword:         os.Getenv("SMTP_PASSWORD"),
		SMTPFromName:         getEnv("SMTP_FROM_NAME", "软考学系"),
		// OPT-04: 生产环境默认不绕过任何用户（必须显式设置 ADMIN_BYPASS_USERNAMES）；
		// 开发环境保持默认 ["ross"] 兼容旧项目初始化逻辑。
		AdminBypassUsernames: loadAdminBypassUsernames(getEnv("APP_ENV", "development")),
		// OPT-01: CORS 白名单
		CORSAllowedOrigins: getEnvCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		// OPT-01: 应用环境，development / production（同时为 OPT-02 准备）
		AppEnv: getEnv("APP_ENV", "development"),
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
	// 打印最终生效的关键配置（密码打码），方便排查 .env 是否被读到
	log.Printf("[config] DB=%s@%s:%s/%s  ServerPort=%s  JWT=%s",
		cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.ServerPort, maskSecret(cfg.JWTSecret))
	return cfg
}

// loadEnvFile 按以下顺序寻找 .env，找到即用：
//  1. ENV_FILE 环境变量指定的绝对路径
//  2. 当前工作目录
//  3. 可执行文件所在目录（systemd / 任意 cwd 启动都能找到）
//  4. /etc/softteststudyt/.env（Linux 标准配置目录）
func loadEnvFile() {
	candidates := []string{}

	if v := os.Getenv("ENV_FILE"); v != "" {
		candidates = append(candidates, v)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}
	if _, err := os.Stat("/etc/softteststudyt/.env"); err == nil {
		candidates = append(candidates, "/etc/softteststudyt/.env")
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				log.Printf("[config] 已加载环境变量文件: %s", p)
				return
			} else {
				log.Printf("[config] 找到 %s 但读取失败: %v", p, err)
			}
		}
	}
	log.Println("[config] 未找到 .env（尝试过: cwd / 可执行目录 / /etc/softteststudyt/.env）" +
		"，将仅依赖进程环境变量")
}

// maskSecret 密钥前 4 后 4，中间打码，避免日志泄露
func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
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

// loadAdminBypassUsernames 加载管理员绕过名单：
//   - 显式设置了 ADMIN_BYPASS_USERNAMES：以环境变量为准
//   - 未设置：development 默认 ["ross"] 保持本地兼容；production 默认空（不绕过任何用户）
func loadAdminBypassUsernames(appEnv string) []string {
	if _, ok := os.LookupEnv("ADMIN_BYPASS_USERNAMES"); ok {
		return getEnvCSV("ADMIN_BYPASS_USERNAMES", nil)
	}
	if appEnv == "production" {
		return []string{}
	}
	return []string{"ross"}
}
