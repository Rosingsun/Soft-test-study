# 软考学系平台 - 配置说明

> 本文档说明后端运行所需的全部环境配置。**真实凭据不得提交到仓库**，所有敏感值仅放在本地 `.env` 中。

## 1. 配置加载机制

后端使用 [`godotenv`](https://github.com/joho/godotenv) 在启动时按以下优先级加载配置：

1. 操作系统环境变量
2. `backend/.env` 文件（仓库 `.gitignore` 已排除）
3. 代码内的默认值（仅用于开发兜底，生产环境必须显式覆盖）

加载逻辑见 `backend/internal/config/config.go:22`。

启动顺序：

```bash
cd backend
go run cmd/server/main.go   # 自动读取 .env
```

## 2. 配置项清单

所有变量定义在 `backend/.env.example`（已纳入版本控制）。字段含义如下：

| 变量 | 必填 | 说明 | 默认值 |
|------|------|------|--------|
| `DB_HOST` | 是 | MySQL 主机地址 | `localhost` |
| `DB_PORT` | 是 | MySQL 端口 | `3306` |
| `DB_USER` | 是 | 数据库用户名 | `root` |
| `DB_PASSWORD` | 是 | 数据库密码 | （无，缺失时启动告警） |
| `DB_NAME` | 是 | 数据库名 | `softteststudyt` |
| `JWT_SECRET` | 是 | JWT 签名密钥，**生产环境必须使用强随机值** | `dev-secret-change-me`（启动会告警） |
| `JWT_EXPIRES_IN` | 否 | Token 有效期（小时） | `168`（7 天） |
| `SERVER_PORT` | 否 | HTTP 服务监听端口 | `8080` |
| `SMTP_HOST` | 否 | SMTP 服务器域名（如 `smtp.qq.com`） | （空，未配置时验证码仅打印到日志） |
| `SMTP_PORT` | 否 | SMTP 端口 | `465` |
| `SMTP_USER` | 否 | 发件邮箱账号 | （空） |
| `SMTP_PASSWORD` | 否 | 发件邮箱授权码（非登录密码） | （空） |
| `SMTP_FROM_NAME` | 否 | 发件人显示名 | `软考学系` |

## 3. 本地初始化步骤

### 3.1 准备环境变量文件

```bash
cd backend
cp .env.example .env
```

按本机实际情况修改 `.env`：

```ini
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=你的本地MySQL密码
DB_NAME=softteststudyt

JWT_SECRET=本地任意非空字符串即可
JWT_EXPIRES_IN=168

SERVER_PORT=8080

# 邮箱验证（可选；不配则验证码仅打印到日志，便于开发联调）
# SMTP_HOST=smtp.qq.com
# SMTP_PORT=465
# SMTP_USER=your-account@qq.com
# SMTP_PASSWORD=your-smtp-authorization-code
# SMTP_FROM_NAME=软考学系
```

### 3.2 初始化数据库

```bash
# 在项目根目录
mysql -u root -p < database.sql
mysql -u root -p softteststudyt < seed_questions.sql
mysql -u root -p softteststudyt < seed_questions_batch2.sql
```

### 3.3 启动后端

```bash
cd backend
go mod download
go run cmd/server/main.go
```

启动成功会看到 `Server listening on :8080` 类似日志。

## 4. 生产环境建议

| 项 | 建议 |
|------|------|
| `JWT_SECRET` | 使用 `openssl rand -hex 32` 生成 64 位随机串，配置到云平台密钥管理（不要写进代码） |
| `DB_PASSWORD` | 创建专用账号并限制来源 IP；定期轮换 |
| 配置注入 | 优先用平台托管的环境变量（Docker / k8s Secret / 阿里云参数仓库），不要把 `.env` 上传 |
| HTTPS | 反向代理（Nginx / Caddy）终止 TLS，后端只监听内网 |

## 5. 敏感信息保护

`.gitignore` 已显式忽略：

```gitignore
.env
*.env.local
```

**注意事项**：

- **禁止**将含真实凭据的 `.env` 提交或推送到任何仓库（含私有仓库外的镜像）
- 提交前可使用 `git status` 确认无 `.env`；或安装 [`gitleaks`](https://github.com/gitleaks/gitleaks) / [`detect-secrets`](https://github.com/Yelp/detect-secrets) 做密钥扫描
- 一旦误提交，**仅删除文件无法消除泄露**，必须立即在源头（数据库、密钥服务）轮换凭据，并清理 Git 历史（`git filter-repo` / BFG）

## 6. 常见问题

**Q：启动报 `警告: 未配置 DB_PASSWORD 环境变量`**
A：`.env` 不存在或 `DB_PASSWORD=` 为空，检查文件路径与字段值。

**Q：连接数据库报 `Access denied for user`**
A：检查 `DB_USER` / `DB_PASSWORD` 是否正确，以及该账号是否拥有 `softteststudyt` 库的权限。

**Q：JWT 一直报 `token invalid`**
A：本地多次重启若改动过 `JWT_SECRET`，旧 token 会失效，重新登录即可。生产环境严禁重启时变更密钥。

**Q：端口被占用**
A：修改 `SERVER_PORT` 后重启；或先 `lsof -i :8080` 找到占用进程。
