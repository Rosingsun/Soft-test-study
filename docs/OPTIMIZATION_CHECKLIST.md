# 软考学系平台 - 优化清单

> 本文档列出项目可优化项，**仅做技术优化，业务功能与业务流程保持不变**。
> 每个任务均有唯一 ID、明确的变更范围与验收标准，可直接分配给 Agent 独立执行。
>
> - 任务状态：`pending` / `in_progress` / `completed` / `cancelled`
> - 优先级：`P0`（上线前必修）/ `P1`（尽快修复）/ `P2`（择期优化）
> - 风险等级：`低` / `中` / `高`（指改动本身引入回归的概率）

---

## 0. 通用约定

执行任何任务前，**必须**先阅读本节。所有任务共同遵守以下约束：

### 0.1 禁止变更的范围

- 业务功能：用户故事、接口的输入输出语义、UI 行为
- 接口契约：`backend/internal/dto/*.go` 与 `frontend/src/api/*.ts` 中已声明的请求/响应字段名、字段顺序、枚举值
- 数据库 schema：表名、列名、外键关系、数据类型（**仅**允许新增索引，不得重命名或删除列）
- 前端组件 API：`<BaseButton>`、`<BaseCard>` 等基础组件对外暴露的 props 与 slots
- 路由路径：`/api/v1/*` 与前端 `router/index.ts` 中声明的路径
- 现有数据库数据：所有数据迁移脚本**必须**保证存量数据可读、可写、行为一致

### 0.2 通用流程

1. 修改前先 `git status` 确认工作区干净
2. 一次只做一个任务，commit 信息格式：`optimize: <任务ID> <任务标题>`
3. 修改后必须执行：
   - 后端：`go build ./...` 与 `go vet ./...` 通过
   - 前端：`npm run typecheck`（若已配置）与 `npm run build` 通过
4. 涉及数据库变更：编写可重入的 migration 文件，文件名递增
5. 修改完成后更新本文档对应任务状态为 `completed`，并在「备注」栏记录 commit hash

### 0.3 任务依赖关系

```
OPT-01 ──┐
OPT-02 ──┤
OPT-04 ──┴──> OPT-08（可并发开发，但 OPT-08 联调时必须 OPT-01~04 已完成）
OPT-18 ──> OPT-19（OPT-18 提供后端批量接口，OPT-19 调用之）
```

---

## 1. P0 任务（上线前必修）

### OPT-01 收紧 CORS 配置

- **优先级**：P0
- **类别**：安全
- **风险等级**：低
- **影响文件**：`backend/cmd/server/main.go`、`backend/internal/middleware/cors.go`（如不存在则新建）
- **状态**：completed
- **Commit**：`65e21bc`

**现状**：
- `cmd/server/main.go:24-35` 的 CORS 中间件对 `Access-Control-Allow-Origin` 返回通配符 `*`，或硬编码允许列表
- 不区分环境，开发与生产共用同一份配置

**变更要求**：
1. 在 `internal/config/config.go` 增加配置项 `CORS_ALLOWED_ORIGINS`（字符串，逗号分隔），默认值为 `http://localhost:5173`
2. CORS 中间件按白名单动态设置 `Access-Control-Allow-Origin`；不在白名单的 origin 返回**无** `Access-Control-Allow-Origin` 头
3. 允许的请求方法限定为 `GET, POST, PUT, DELETE, PATCH, OPTIONS`
4. 允许的请求头限定为 `Origin, Content-Type, Authorization, X-Requested-With`
5. `Access-Control-Allow-Credentials` 仅在 origin 命中白名单时设为 `true`，否则不设置
6. 生产环境（`APP_ENV=production`）若 `CORS_ALLOWED_ORIGINS` 为空，启动时 `log.Fatal` 拒绝启动

**禁止变更**：
- 不修改 CORS 中间件的调用位置与注册顺序
- 不修改路由的 `Method` 列表

**验收标准**：
- `curl -H "Origin: http://evil.com" http://localhost:8080/api/v1/subjects -i` 响应中**无** `Access-Control-Allow-Origin` 头
- `curl -H "Origin: http://localhost:5173" http://localhost:8080/api/v1/subjects -i` 响应头包含 `Access-Control-Allow-Origin: http://localhost:5173`
- `go build ./...` 通过

---

### OPT-02 JWT 密钥生产环境强制校验

- **优先级**：P0
- **类别**：安全
- **风险等级**：低
- **影响文件**：`backend/internal/config/config.go`、`backend/cmd/server/main.go`
- **状态**：completed
- **Commit**：`4ec6999`

**现状**：
- `config.go:39` `JWT_SECRET` 默认值为 `dev-secret-change-me`
- 启动时仅 `log.Printf` 警告，未阻止生产环境启动

**变更要求**：
1. 在 `config.Config` 中增加字段 `AppEnv`（字符串，从 `APP_ENV` 读取，默认 `development`）
2. 启动时（`main.go` `LoadConfig` 之后）增加校验：
   - 当 `AppEnv == "production"` 时，若 `JWT_SECRET` 为空、为 `dev-secret-change-me`、或长度小于 32 字节，调用 `log.Fatal` 终止启动，错误信息为 `JWT_SECRET 未配置或强度不足，禁止生产环境启动`
3. 在 `.env.example` 中将 `APP_ENV` 加入并附注释：`# development | production`

**禁止变更**：
- 不修改 JWT 签发逻辑（`internal/middleware/auth.go`）
- 不修改 token 有效期默认值
- 不修改任何业务接口对 token 的使用方式

**验收标准**：
- `APP_ENV=production JWT_SECRET=dev-secret-change-me go run cmd/server/main.go` 进程退出码非 0，日志包含 `JWT_SECRET 未配置`
- `APP_ENV=production JWT_SECRET=$(openssl rand -hex 32) go run cmd/server/main.go` 启动成功
- `APP_ENV=development JWT_SECRET=dev-secret-change-me go run cmd/server/main.go` 启动成功（保持向后兼容）

---

### OPT-03 修复静态文件路径穿越漏洞

- **优先级**：P0
- **类别**：安全
- **风险等级**：中
- **影响文件**：`backend/internal/service/study_material.go:77-84`、`backend/cmd/server/main.go`
- **状态**：completed
- **Commit**：`e2fe15a`

**现状**：
- `cmd/server/main.go:44` 使用 `r.Static("/uploads", "./data/uploads")` 暴露整个 uploads 目录
- `service/study_material.go:77-84` `resolveFilePath` 将 `fileURL`（用户可控字段）直接 `strings.TrimPrefix` 后拼接到 `data/uploads/`，**未校验是否仍在该目录下**
- 攻击者上传文件名包含 `../` 即可读取服务器任意文件

**变更要求**：
1. 在 `internal/service/study_material.go` 增加函数 `resolveFilePath(fileURL string) (string, error)`：
   - 解析 `fileURL` 为相对路径后，调用 `filepath.Clean`
   - 计算 `baseDir` 的绝对路径（`filepath.Abs`）
   - 计算最终路径的绝对路径，必须以 `baseDir` 为前缀（`strings.HasPrefix`），否则返回 `errors.New("非法文件路径")`
2. 删除 `cmd/server/main.go:44` 的 `r.Static` 调用
3. 在 `internal/handler/study_material.go` 增加路由 `GET /api/v1/study-materials/:id/file`：
   - 从 path 取 `id`，调用 `service.GetByID(id, userID)` 获取记录
   - 鉴权：仅 `uploader_id == userID` 或 `user_role == "admin"` 可下载
   - 调用 `resolveFilePath` 解析路径，失败返回 403
   - 设置 `Content-Disposition: attachment; filename="<原始文件名>"` 与对应 `Content-Type`
4. 保留 `/uploads/*` 路由但**仅**用于上传文件落盘使用，不注册到 HTTP handler（即 `main.go` 不挂载该路径）

**禁止变更**：
- 不修改上传接口的请求/响应字段
- 不修改 `study_materials` 表结构
- 不修改 `file_url` 字段的存储格式

**验收标准**：
- 数据库中某条 `file_url = "/uploads/../../etc/passwd"`，调用下载接口返回 403
- 正常 `file_url = "/uploads/xxx.pdf"`，下载成功，响应头包含 `Content-Disposition: attachment`
- 静态目录不再被外部 HTTP 访问（`curl http://localhost:8080/uploads/` 返回 404）
- `go build ./...` 通过

---

### OPT-04 新增 Admin 角色中间件

- **优先级**：P0
- **类别**：安全
- **风险等级**：中
- **影响文件**：`backend/internal/middleware/`（新建 `admin.go`）、`backend/internal/router/router.go`
- **状态**：completed
- **Commit**：`e481b34`

**现状**：
- 后端无任何 admin 角色校验中间件
- 前端 `router/index.ts` 通过 `meta.role = 'admin'` 隐藏路由，但绕过前端直接调用 API 仍可访问
- `ADMIN_BYPASS_USERNAMES`（默认含 `ross`）注册时直接授予 admin 角色

**变更要求**：
1. 在 `internal/middleware/` 新建 `admin.go`：
   - 函数 `RequireAdmin() gin.HandlerFunc`
   - 从 `c.Get("user_id")` 取 `uint` 用户 ID
   - 通过 `repository.UserRepo.GetByID` 查用户角色
   - 角色不等于 `admin` 返回 403 + `{"code": 10003, "message": "无权限"}`
2. 在 `internal/router/router.go` 中给以下路由组加 `RequireAdmin()`：
   - `/api/v1/admin/*`（所有 admin 路由）
   - `/api/v1/ai/import-real`、`/api/v1/ai/import-case-study`（题目导入）
   - `/api/v1/ai/syllabus`、`/api/v1/ai/syllabus-questions`（教学大纲管理）
   - `/api/v1/ai/batches/:batch_id/approve`（题目审批）
   - 任何**新**的 `/api/v1/admin/*` 路由
3. 在 `internal/config/config.go` 增加配置项 `AdminBypassUsernames` 读取 `ADMIN_BYPASS_USERNAMES`，默认值为空字符串（即**生产环境**不绕过任何用户）
4. 现有注册 bypass 逻辑（`service/user.go`）改为：仅当 `config.AdminBypassUsernames` 包含该用户名时授予 admin 角色

**禁止变更**：
- 不修改 `users` 表结构
- 不修改 `roles` 字段的取值集合（`user` / `admin`）
- 不修改普通用户的注册、登录逻辑
- 不修改 `ADMIN_BYPASS_USERNAMES` 的环境变量名（保持向后兼容）

**验收标准**：
- 数据库中 `roles='user'` 的用户访问任意 admin 路由返回 403
- 数据库中 `roles='admin'` 的用户访问 admin 路由正常
- `ADMIN_BYPASS_USERNAMES=ross` 时注册 `ross` 仍可获得 admin（开发兼容）
- `APP_ENV=production ADMIN_BYPASS_USERNAMES=` 时注册 `ross` 角色为 `user`

---

### OPT-05 AI API Key 存储安全

- **优先级**：P0
- **类别**：安全
- **风险等级**：低
- **影响文件**：`frontend/src/stores/ai.ts`、`frontend/src/api/ai.ts`
- **状态**：completed
- **Commit**：`9b860b8`

**现状**：
- `stores/ai.ts:27-29` 将用户填入的 AI API Key 明文存入 `localStorage`
- XSS 漏洞即可窃取所有用户 AI Key

**变更要求**：
1. `stores/ai.ts` 中：
   - 增加 `aiKey` 状态字段
   - `setAiKey(key: string)` 时：若非空则 `sessionStorage.setItem('ai_key', key)`；若为空则清除
   - 提供 `getAiKey(): string` 从 `sessionStorage` 读取
   - 提供 `clearAiKey()` 清除
   - 监听 `auth/logout` 事件（监听 Pinia 的 `auth` store `$onAction` 或自定义事件），自动调用 `clearAiKey`
2. `api/ai.ts` 中调用方从 `sessionStorage` 实时取 Key，不再从 store state 取
3. 增加 `app/storage.ts` 工具模块集中管理 sessionStorage 操作（`getItem` / `setItem` / `removeItem`），后续其他敏感字段也走此模块

**禁止变更**：
- 不修改 AI 接口的请求/响应格式
- 不修改后端对 Key 的处理逻辑
- 不删除 `localStorage` 中已存在的旧 Key 数据（保持向后兼容：首次迁移时若发现 `localStorage.ai_key`，迁移到 `sessionStorage` 后删除 localStorage 项）

**验收标准**：
- 用户在设置页填入 Key → 关闭浏览器标签页后重新打开 → Key 不存在（sessionStorage 失效）
- 浏览器 DevTools → Application → Local Storage 中**不**包含 `ai_key` 键
- 登录状态下正常使用 AI 功能（出题、分析）不受影响
- 调用 `logout` 后 Key 被清除
- `npm run typecheck` 通过

---

### OPT-06 配置可信代理白名单

- **优先级**：P0
- **类别**：安全
- **风险等级**：低
- **影响文件**：`backend/cmd/server/main.go`
- **状态**：completed
- **Commit**：`c162044`

**现状**：
- `cmd/server/main.go` 调用 `gin.Default()` 后未调用 `r.SetTrustedProxies()`
- 反向代理场景下 `c.ClientIP()` 读取 `X-Forwarded-For` 头，**攻击者可伪造**
- 限流、日志中的客户端 IP 全部失真

**变更要求**：
1. 在 `internal/config/config.go` 增加配置项 `TrustedProxies`（字符串，逗号分隔），从 `TRUSTED_PROXIES` 读取，默认 `127.0.0.1,::1`
2. `main.go` 在 `r := gin.New()` 之后调用 `r.SetTrustedProxies(strings.Split(cfg.TrustedProxies, ","))`
3. 在 `.env.example` 增加 `TRUSTED_PROXIES=127.0.0.1,::1` 并附注释：`# 反向代理 IP，多个用逗号分隔，公网部署需包含 Nginx 网段`

**禁止变更**：
- 不修改任何业务接口中对 `c.ClientIP()` 的调用
- 不修改 `gin.Recovery()` / `gin.Logger()` 的默认行为

**验收标准**：
- `curl -H "X-Forwarded-For: 1.2.3.4" http://localhost:8080/api/v1/subjects` 启动日志（如果打印了 IP）显示为 `127.0.0.1`（loopback）
- `TRUSTED_PROXIES=192.168.1.10` 时，相同 curl 显示为 `1.2.3.4`
- `go build ./...` 通过

---

### OPT-07 修复 StartExam 并发重复创建

- **优先级**：P0
- **类别**：安全 / 数据一致性
- **风险等级**：中
- **影响文件**：`backend/internal/service/exam.go:88-150`、`backend/internal/migrations/sql/`（新增 migration）
- **状态**：completed
- **Commit**：`42c8467`

**现状**：
- `StartExam` 检查 pending → 创建 record 期间无事务、无锁
- 同一用户对同一模板的并发请求可创建多条 pending 记录
- 后果：扣分/答题数据错乱

**变更要求**：
1. 新建 migration `000XXX_unique_pending_exam_record.up.sql`（XXX 取当前最大序号 + 1）：
   ```sql
   -- 删除历史脏数据（保留每组最新一条）
   DELETE r1 FROM exam_records r1
   INNER JOIN exam_records r2
   ON r1.user_id = r2.user_id
   AND r1.template_id = r2.template_id
   AND r1.status = 'pending' AND r2.status = 'pending'
   AND r1.id < r2.id;
   
   -- 新增部分唯一索引（MySQL 8.0+）
   CREATE UNIQUE INDEX uk_exam_pending
   ON exam_records (user_id, template_id, status)
   WHERE status = 'pending';
   ```
2. `service/exam.go` 的 `StartExam`：
   - 用 `s.db.Transaction` 包裹「查询 pending → 创建 record」
   - 创建失败时捕获 MySQL 错误码 1062，翻译为业务错误 `已有进行中的考试`
   - 返回该错误时 `handler` 层映射为 `code=10009, http_status=409`

**禁止变更**：
- 不修改 `exam_records` 表的其他列
- 不修改 `StartExam` 的请求/响应结构
- 不修改 `exam_records.status` 的取值集合

**验收标准**：
- 启动后 `SHOW INDEX FROM exam_records` 出现 `uk_exam_pending`
- 数据库存在脏数据时 migration 自动清理（执行前后 `SELECT COUNT(*) FROM exam_records WHERE status='pending' GROUP BY user_id, template_id HAVING COUNT(*)>1` 结果为 0）
- 并发 10 次同用户同模板 `StartExam` 请求，仅 1 次成功，其余 9 次返回 409
- 已完成（status='completed'）的考试不参与唯一约束
- `go build ./...` 与 `go test ./...`（如有）通过

---

### OPT-08 AI 分析/评分接口鉴权

- **优先级**：P0
- **类别**：安全
- **风险等级**：中
- **影响文件**：`backend/internal/handler/ai.go:218, 312`
- **状态**：completed
- **Commit**：`986bfd3`

**现状**：
- `Analyze`（约 218 行）与 `CheckEssayScore`（约 312 行）handler 未挂载任何鉴权中间件
- 任何匿名请求可调用并消耗 LLM token / 窃取他人 essay 评分

**变更要求**：
1. `Analyze` 接口：
   - 在 `router/router.go` 中将 `/ai/analyze` 加入 `auth` 中间件组
   - `handler` 中从 `c.Get("user_id")` 取用户 ID，作为 `req.UserID` 字段传入 service（若无则先在 dto 增加可选 `user_id` 字段，后端强制覆盖为 token 中的 ID）
2. `CheckEssayScore` 接口：
   - 在 `router/router.go` 中将 `/ai/essay-score` 加入 `auth` 中间件组
   - `handler` 中先 `recordID := c.Param("id")` → service 层校验 `record.user_id == currentUserID`，不匹配返回 403
3. `ai.go` 中其他**无鉴权**的 handler（`GetBatchQuestions`、`GenerateQuestions` 等）一律改为必须登录

**禁止变更**：
- 不修改 AI 接口的请求/响应字段（仅在 server 端从 token 覆盖 userID，前端不需要传）
- 不修改 LLM 调用逻辑与 prompt 模板
- 不修改计费/审计逻辑（如果存在）

**验收标准**：
- 未携带 token 调用 `/api/v1/ai/analyze` 返回 401
- 携带普通用户 token 调用 `CheckEssayScore` 但 `record.user_id` 不等于自己 → 返回 403
- 携带 admin token 调用任意 AI 接口正常
- `go build ./...` 通过

---

## 2. P1 任务（尽快修复）

### OPT-09 限流器替换与内存泄漏修复

- **优先级**：P1
- **类别**：性能 / 内存
- **风险等级**：中
- **影响文件**：`backend/internal/middleware/ratelimit.go`

**现状**：
- 当前限流使用 `map[string]*visitor` 存储访问记录
- 访问记录**永不清理**，进程长时间运行后内存持续增长
- 进程重启即清零，多实例部署无法共享限流状态

**变更要求**：
1. 使用 `golang.org/x/time/rate` 实现 IP 级别的 token bucket：
   - 每个 IP 维护一个 `*rate.Limiter`
   - 通用阈值：60 次/分钟
   - 登录、注册、发送验证码等接口：10 次/分钟
2. 用 `sync.Map` 或 `map + sync.RWMutex` 存储 `*rate.Limiter` 实例
3. 增加**后台 goroutine**：每 10 分钟遍历 map，若某 IP 最近 1 小时无访问则删除其 `*rate.Limiter`（使用 `atomic.Int64` 记录最后访问时间）
4. **不**引入 Redis 依赖（保持单进程方案）

**禁止变更**：
- 不修改 `ratelimit.go` 对外暴露的函数签名
- 不修改调用方传入的 key 计算逻辑（如 IP 取值方式）
- 不修改限流错误码（保持 `10004` 不变，避免与 `CodeForbidden` 冲突，参考 OPT-23）

**验收标准**：
- 1 分钟内 70 次同 IP 登录请求，前 10 次通过，后 60 次返回 429
- 进程运行 24 小时后 `pprof heap` 中限流 map 节点数 < 1000
- `go test` 中包含 1 个针对限流器的单元测试（mock clock）

---

### OPT-10 写接口全局限流

- **优先级**：P1
- **类别**：安全 / 性能
- **风险等级**：低
- **影响文件**：`backend/internal/router/router.go`、`backend/internal/middleware/ratelimit.go`

**现状**：
- 限流仅覆盖 `/auth/login`、`/auth/register`、`/auth/email/send-code`、`/ai/*`
- `/practice/submit`、`/exam-records/:id/submit-answer`、`/wrong-questions/practice` 等高频写接口**无任何限流**

**变更要求**：
1. 在 `router/router.go` 增加中间件 `WriteRateLimit()`，对所有 `POST/PUT/DELETE/PATCH` 方法应用，阈值 120 次/分钟/IP
2. 用户级（基于 `user_id`）写接口限流：300 次/分钟，在 `auth` 中间件之后挂载
3. 写接口路由组的中间件链：`cors → ratelimit(IP) → auth → ratelimit(user)`

**禁止变更**：
- 不修改已有 `/auth/*` 限流逻辑
- 不修改 GET 请求的处理

**验收标准**：
- 同 IP 1 分钟内发起 200 次 POST `/practice/submit`，后 80 次返回 429
- 多个写接口叠加不超过 300 次/分钟
- `go build ./...` 通过

---

### OPT-11 bcrypt 成本提升

- **优先级**：P1
- **类别**：安全
- **风险等级**：低
- **影响文件**：`backend/internal/service/user.go`

**现状**：
- `bcrypt.GenerateFromPassword` 使用 `bcrypt.DefaultCost`（cost=10）
- 现代 GPU 每秒可穷举 10^4 次，弱密码在数小时内可被破解

**变更要求**：
1. 将 `bcrypt.DefaultCost` 替换为常量 `bcryptCost = 12`
2. 在 `service/user.go` 顶部声明 `const bcryptCost = 12`
3. 在 `pkg/security/password.go`（如不存在则新建）封装 `HashPassword(p string) (string, error)` 与 `VerifyPassword(hash, p string) error`，内部使用 `bcryptCost`
4. 全局替换 `bcrypt.GenerateFromPassword` 与 `bcrypt.CompareHashAndPassword` 调用为上述封装函数

**禁止变更**：
- 不修改密码字段存储格式（仍存 bcrypt hash 字符串）
- 不修改 `users.password` 列类型
- 旧 hash（cost=10）**必须**仍能正常登录（`CompareHashAndPassword` 自动识别 cost）

**验收标准**：
- 新注册用户，数据库中密码 hash 以 `$2a$12$` 开头
- 已有 cost=10 的账号登录正常
- 100 次登录平均响应时间增加不超过 300ms
- `go test` 通过

---

### OPT-12 修复登录锁定的并发安全问题

- **优先级**：P1
- **类别**：安全 / 数据一致性
- **风险等级**：中
- **影响文件**：`backend/internal/service/user.go:146-236, 260-340`

**现状**：
- 登录失败 5 次锁定 15 分钟，但：
  - 锁定状态在内存 model 中读写，**无并发安全**（多个并发登录请求同时判定）
  - 锁定期间 `FailedAttempts` 被重置为 0，下一轮又重新计 5 次
  - 锁定时长写死 15 分钟，无配置项

**变更要求**：
1. 在 `users` 表**不新增列**（保持 schema 不变），改为：
   - `LockedUntil` 字段（`time.Time`）已有则复用；如无则保持现状，改用 `FailedAttempts` 字段语义：
     - 失败 5 次后 `FailedAttempts=99`（哨兵值），记录当前时间到内存 `map[uint]time.Time` 中
2. 改为线程安全：
   - 用 `sync.Map` 或 `map + sync.Mutex` 维护 `lockedUsers map[uint]time.Time`
   - 判定时先获取锁再检查
3. 在 `config.go` 增加配置项：
   - `LoginMaxFailed = 5`（默认）
   - `LoginLockMinutes = 15`（默认）
4. 启动时加载配置

**禁止变更**：
- 不修改 `users` 表的列
- 不修改登录接口的请求/响应字段
- 不修改前端登录失败提示文案（错误信息仍为 `用户名或密码错误`）

**验收标准**：
- 同用户 5 次错误密码后第 6 次返回 `账号已锁定，请 15 分钟后再试`
- 锁定 15 分钟后可正常登录
- 并发 10 次同用户错误登录，仅 1 次返回锁定提示，其余 9 次正常 `用户名或密码错误`
- `go build ./...` 通过

---

### OPT-13 学习计划 N+1 查询优化

- **优先级**：P1
- **类别**：性能
- **风险等级**：中
- **影响文件**：`backend/internal/service/study_plan.go:42-50, 167-217`

**现状**：
- `toResp` 每个计划循环调用 `CountOnDate` + `CountInRange`
- 5 个计划 → 10 次数据库 RTT

**变更要求**：
1. 在 `repository/study_plan.go` 新增方法 `BatchStats(userID uint, planIDs []uint, date time.Time) (map[uint]DailyStats, error)`：
   - 单条 SQL：`SELECT plan_id, SUM(...), COUNT(...) FROM check_ins WHERE user_id=? AND plan_id IN (?) AND check_date BETWEEN ? AND ? GROUP BY plan_id`
2. `toResp` 改为：
   - 先 `BatchStats` 一次性取所有计划的统计数据
   - 循环中从 map 取值，不再调单条查询

**禁止变更**：
- 不修改 `study_plans` / `check_ins` 表结构
- 不修改 `StudyPlanResp` dto 的字段
- 不修改 `toResp` 对外签名

**验收标准**：
- 用户有 5 个学习计划时，`ListPlans` 接口 SQL 日志仅出现 1 条聚合查询
- 接口 P99 响应时间降低 ≥ 50%
- 现有功能（计划展示、连续打卡天数计算）行为完全一致
- `go build ./...` 与 `go test` 通过

---

### OPT-14 错题本/收藏页串行加载优化

- **优先级**：P1
- **类别**：性能
- **风险等级**：低
- **影响文件**：`backend/internal/repository/question.go`（或新建 `question_batch.go`）、`backend/internal/handler/question.go`、前端 `frontend/src/views/practice/WrongPractice.vue`、`frontend/src/views/practice/Favorites.vue`

**现状**：
- `WrongPractice.vue:24-37` 与 `Favorites.vue:49-54` 串行 `await getQuestion(id)` 拉取题目详情
- 20 道错题 → 20 次串行请求

**变更要求**：
1. 后端：
   - 新建 `repository.BatchGetByIDs(ids []uint) ([]model.Question, error)`，单条 SQL：`SELECT * FROM questions WHERE id IN (?) AND deleted_at IS NULL`
   - `handler/question.go` 新增路由 `POST /api/v1/questions/batch`，请求体 `{"ids": [1,2,3]}`，响应 `[]QuestionResp`
2. 前端：
   - `api/question.ts` 新增 `batchGetQuestions(ids: number[]): Promise<QuestionResp[]>`
   - `WrongPractice.vue` / `Favorites.vue` 改为单次 `batchGetQuestions` 调用
3. dto 与 types 同步更新（参考 `docs/API.md` 中的字段命名规范）

**禁止变更**：
- 不修改单题详情接口 `GET /questions/:id` 的行为
- 不修改 `QuestionResp` 字段
- 不修改前端错题本/收藏页的 UI 与交互

**验收标准**：
- 加载 20 道错题时 Network 面板**只出现 1 个 `/questions/batch` 请求**
- 加载耗时从 `20 × RTT` 降为 `1 × RTT`
- `npm run typecheck` 与 `go build ./...` 通过

---

### OPT-15 数据库索引补齐

- **优先级**：P1
- **类别**：性能
- **风险等级**：低
- **影响文件**：`backend/internal/migrations/sql/`（新增 migration）

**现状**：
- `notifications` 仅 `idx_user_id`、`idx_read`，组合查询效率低
- `wrong_questions` 排序 `wrong_count DESC` 走 filesort
- `email_verification_codes` 缺高频查询复合索引
- `ai_generated_questions.question_id` 无 FK

**变更要求**：
1. 新建 migration `000XXX_add_indexes.up.sql`：
   ```sql
   CREATE INDEX idx_notif_user_read_created
     ON notifications (user_id, read, created_at DESC);
   
   CREATE INDEX idx_wrong_user_count
     ON wrong_questions (user_id, wrong_count DESC);
   
   CREATE INDEX idx_email_user_purpose_used
     ON email_verification_codes (user_id, purpose, used, created_at DESC);
   
   ALTER TABLE ai_generated_questions
     ADD CONSTRAINT fk_aiq_question
     FOREIGN KEY (question_id) REFERENCES questions(id)
     ON DELETE SET NULL;
   ```
2. 对应 `.down.sql` 文件

**禁止变更**：
- 不删除任何已有索引
- 不修改任何列定义

**验收标准**：
- `EXPLAIN SELECT * FROM notifications WHERE user_id=1 AND read=0 ORDER BY created_at DESC` 走 `idx_notif_user_read_created`
- 三个查询的 `type` 列为 `ref` 或 `range`，`Extra` 不含 `filesort`
- 存量数据保持不变
- `go build ./...` 通过

---

### OPT-16 考试记录与错题列表分页

- **优先级**：P1
- **类别**：性能
- **风险等级**：中
- **影响文件**：`backend/internal/service/exam.go:511-567`、`backend/internal/service/wrong_question.go:29-82`、`backend/internal/handler/exam.go:99-106`

**现状**：
- `ListRecords`、`ListWrongQuestions` 一次返回所有记录
- 重度用户 100+ 条考试记录时，前端渲染与带宽成瓶颈

**变更要求**：
1. `dto` 增加通用分页请求 `PageReq{Page int, PageSize int}`：
   - `Page` 默认 1，最小 1
   - `PageSize` 默认 20，最大 100
2. service 层两个 `List` 方法签名改为接受 `PageReq`，返回 `(items []T, total int64, err error)`
3. handler 层从 query string 解析 `page` / `page_size`，响应使用 `response.SuccessPage(c, items, total, page, pageSize)`
4. 前端 `views/exam/ExamHistory.vue`、`views/practice/WrongQuestions.vue` 增加分页参数与翻页器

**禁止变更**：
- 不修改单条记录的响应结构
- 不修改 `ListRecords` 已有的查询条件（科目、状态、时间范围）
- 不修改 `WrongQuestions` 列表的展示字段

**验收标准**：
- 请求 `?page=1&page_size=20` 返回前 20 条 + `total` 字段
- 请求 `?page=2&page_size=20` 返回第 21-40 条
- `page_size=200` 被截断为 100
- 前端翻页器正常切换，URL 同步 query
- `go build ./...` 与 `npm run typecheck` 通过

---

### OPT-17 Vite 打包体积优化

- **优先级**：P1
- **类别**：性能
- **风险等级**：低
- **影响文件**：`frontend/vite.config.ts`
- **状态**：completed
- **Commit**：`9273523`

**现状**：
- 未配置 `build.rollupOptions.output.manualChunks`
- `echarts` 等大依赖进首屏 vendor chunk

**变更要求**：
1. `vite.config.ts` 增加：
   ```ts
   build: {
     chunkSizeWarningLimit: 1000,
     rollupOptions: {
       output: {
         manualChunks: {
           vue: ['vue', 'vue-router', 'pinia'],
           echarts: ['echarts', 'echarts/vue'],
         }
       }
     }
   }
   ```
2. 根据实际 `package.json` 中的依赖调整 chunks 列表
3. 增加 `build --analyze` 命令（可选）通过 `rollup-plugin-visualizer` 生成报告

**禁止变更**：
- 不修改任何业务代码
- 不修改 `index.html` 入口

**验收标准**：
- `npm run build` 产物中 `echarts` 相关 chunk 与 `vue` chunk 分离
- 首屏（`/` 路由）加载 chunk 总大小（gzipped）减少 ≥ 20%
- `npm run build` 成功无报错

---

### OPT-18 NotificationService 层级重构

- **优先级**：P1
- **类别**：质量 / 规范
- **风险等级**：中
- **影响文件**：`backend/internal/service/notification.go`、`backend/internal/repository/notification.go`（新建）
- **状态**：completed
- **Commit**：`1e36a7c`

**现状**：
- `NotificationService` 直接持有 `*gorm.DB`，绕过 repository 层
- 违反 AGENTS.md「repository 层做数据库查询」的分层规则

**变更要求**：
1. 新建 `repository/notification.go`：
   - 结构体 `NotificationRepo` 持有 `*gorm.DB`
   - 迁移所有 service 层中 `s.db.Where/Find/Create/...` 调用
   - 公开方法：`List`, `GetByID`, `Create`, `MarkRead`, `MarkAllRead`, `UnreadCount`, `Delete`
2. `service/notification.go` 改为持有 `*repository.NotificationRepo`（与其他 service 一致）
3. 构造函数 `NewNotificationService(repo *repository.NotificationRepo) *NotificationService`

**禁止变更**：
- 不修改 service 对外暴露的方法签名
- 不修改 `notifications` 表结构
- 不修改 service 内部业务逻辑（如通知创建时机）

**验收标准**：
- `service/notification.go` 中**不出现** `s.db.` 任何调用
- 现有 `notification` 相关接口行为完全一致
- `go build ./...` 通过

---

### OPT-19 AI 任务管理依赖注入重构

- **优先级**：P1
- **类别**：质量 / 规范
- **风险等级**：中
- **影响文件**：`backend/internal/service/ai_task.go`、`backend/internal/service/ai.go`、`backend/cmd/server/main.go`
- **状态**：completed
- **Commit**：`7476ef2`

**现状**：
- 包级全局变量 `aiTaskStore` / `aiTaskRepo` / `aiTaskTimeoutNotifier`
- 启动顺序敏感，不可单元测试

**变更要求**：
1. 新建 `service/ai_task_manager.go`：
   - 结构体 `AiTaskManager` 持有 store / repo / notifier
   - 公开方法 `New(store, repo, notifier) *AiTaskManager`
2. 删除所有包级 `var aiTask...`
3. 删除 `SetAITaskRepo` / `RegisterTimeoutNotifier` 函数
4. `AiService` 改为持有 `*AiTaskManager`，通过构造函数注入
5. `main.go` 中按依赖顺序构造：`store → repo → manager → aiService`

**禁止变更**：
- 不修改 AI 任务相关的接口（`/ai/tasks/*`）请求/响应
- 不修改 LLM 调用逻辑与 prompt 模板
- 不修改任务超时时间（默认 10 分钟）

**验收标准**：
- `grep -r "package-level\|aiTaskStore" internal/service/ai_task.go` 结果为空
- 启动后 AI 任务创建、查询、审批功能完全正常
- `go build ./...` 通过

---

### OPT-20 Janitor goroutine 增加 context 取消

- **优先级**：P1
- **类别**：质量
- **风险等级**：低
- **影响文件**：`backend/internal/service/exam.go:100, 255, 619`、`backend/cmd/server/main.go`
- **状态**：completed
- **Commit**：`d9406a8`

**现状**：
- 多个 janitor 后台 goroutine（清理过期考试记录、清理 pending 答题等）使用 `go func() { for range ticker.C {...} }`
- 服务关闭时 goroutine 永不退出，**goroutine 泄漏**

**变更要求**：
1. `cmd/server/main.go` 在启动时创建 `ctx, cancel := context.WithCancel(context.Background())`
2. `signal.NotifyContext` 监听 SIGTERM/SIGINT，触发时 `cancel()`
3. 所有 janitor 函数签名改为 `StartJanitorXxx(ctx context.Context)`
4. 内部循环：
   ```go
   select {
   case <-ctx.Done():
       return
   case <-ticker.C:
       // 业务逻辑
   }
   ```
5. 调用方在 main 中 `go StartJanitorCleanupExpiredRecords(ctx)`

**禁止变更**：
- 不修改 janitor 内部的清理逻辑（清理规则、清理范围）
- 不修改 janitor 的执行频率（默认 1 小时）

**验收标准**：
- 服务运行 1 小时后发送 SIGTERM，主进程在 5 秒内退出
- `pprof goroutine` 显示 janitor goroutine 已退出
- `go build ./...` 通过

---

### OPT-21 后端单元测试骨架

- **优先级**：P1
- **类别**：可维护性
- **风险等级**：低
- **影响文件**：`backend/internal/service/`、`backend/internal/middleware/`
- **状态**：completed
- **Commit**：`8b9fa9c`

**现状**：
- 后端无任何 `_test.go` 文件
- 关键判分、SM-2 算法、JWT 生成解析等核心函数无覆盖

**变更要求**：
1. 为以下函数添加 `_test.go` 单元测试（不依赖 DB，使用 `testing` 标准库 + 表驱动测试）：
   - `service/judge.go` 的 `JudgeAnswer(q, ans)` 判分函数
   - `service/sm2.go`（或同功能）SM-2 间隔重复算法
   - `middleware/auth.go` 的 JWT 生成与解析（使用相同 secret）
   - `service/user.go` 的 `VerifyPassword` 与 `HashPassword`
2. 每个 `_test.go` 至少包含 3 个用例（正常 / 边界 / 异常）
3. 在 `package.json`（或后端）配置 `go test ./internal/service/... -v` 脚本

**禁止变更**：
- 不修改被测函数的实现
- 不引入除标准库外的新依赖

**验收标准**：
- `go test ./... -v` 全部通过
- 覆盖率报告 `go test -cover ./internal/service/...` 对上述 4 个文件覆盖率 ≥ 70%

---

### OPT-22 前端 401 处理改为 router 跳转

- **优先级**：P1
- **类别**：质量
- **风险等级**：低
- **影响文件**：`frontend/src/api/request.ts`
- **状态**：completed
- **Commit**：`b11d28e`

**现状**：
- `request.ts:25-32` token 失效时使用 `window.location.href = '/login'` 硬跳转
- 丢失 SPA 状态、当前页面滚动位置

**变更要求**：
1. `request.ts` 改为：从 `vue-router` 注入 router 实例（通过 `main.ts` 设置 `request.setRouter(router)`）
2. 401 时调用 `router.push({ name: 'Login', query: { from: router.currentRoute.value.fullPath } })`
3. `Login.vue` 在挂载时检查 `route.query.from`，登录成功后 `router.push(route.query.from as string)` 回跳
4. 多个并发请求同时 401 时，仅触发一次跳转（用全局标志位）

**禁止变更**：
- 不修改后端 401 响应格式
- 不修改其他 HTTP 状态码的处理
- 不修改登录页 UI

**验收标准**：
- 在任意页面 token 失效后跳转到 `/login?from=/practice`，登录成功后回到 `/practice`
- 跳转让当前页面状态不丢失（虽然页面本身会重渲染，但 URL 保留）
- 5 个并发请求同时 401 仅触发 1 次跳转
- `npm run typecheck` 通过

---

## 3. P2 任务（择期优化）

### OPT-23 修复错误码冲突

- **优先级**：P2
- **类别**：规范
- **风险等级**：低
- **影响文件**：`backend/internal/middleware/ratelimit.go:35`、`backend/internal/config/errors.go`

**现状**：
- 限流错误码 `10004` 与 `config.CodeForbidden=10004` 冲突
- 前端按 code 区分错误时会误判

**变更要求**：
1. 在 `config/errors.go` 增加 `CodeTooManyRequests = 10006`
2. 限流中间件改用 `CodeTooManyRequests`，HTTP 状态码 429
3. 前端 `utils/error.ts` 增加 `code === 10006` 的统一 toast 文案 `操作过于频繁，请稍后再试`

**禁止变更**：
- 不修改其他已有错误码
- 不修改 10001~10005 区间

**验收标准**：
- 触发限流时响应 `{"code": 10006, "message": "..."}` 与 HTTP 429
- 前端 toast 显示「操作过于频繁」
- `go build ./...` 与 `npm run typecheck` 通过

---

### OPT-24 ranking 模块 O(n²) 优化

- **优先级**：P2
- **类别**：性能
- **风险等级**：低
- **影响文件**：`backend/internal/service/ranking.go:53-58`

**现状**：
- `sort.Slice` 后用 O(n) 遍历找"我的排名"，n 较小时可接受，n=10000 时 O(n²) 不可接受

**变更要求**：
1. 排序完成后，用 `sort.Search(len(participants), func(i int) bool { ... })` 二分定位
2. 改用 `slices.IndexFunc` 或类似 O(n) 但更清晰的方法（Go 1.21+）

**禁止变更**：
- 不修改 ranking 接口请求/响应结构
- 不修改分数计算公式

**验收标准**：
- `go test` 包含 1 个 10000 用户的 benchmark
- 优化后 `BenchmarkFindMyRank` 耗时降低 ≥ 50%

---

### OPT-25 静态目录与上传路径使用绝对路径

- **优先级**：P2
- **类别**：质量
- **风险等级**：低
- **影响文件**：`backend/cmd/server/main.go`

**现状**：
- `r.Static("/uploads", "./data/uploads")` 依赖 CWD
- 部署时若 CWD 不是项目根，路径解析失败

**变更要求**：
1. 在 `config.go` 增加 `DataDir`（默认 `./data`）
2. `main.go` 改为：
   ```go
   uploadDir := filepath.Join(cfg.DataDir, "uploads")
   os.MkdirAll(uploadDir, 0755)
   // 不再使用 r.Static，参考 OPT-03 改为 handler 流式返回
   ```

**禁止变更**：
- 不修改 `.env` 中 `DATA_DIR`（如无则新增，**默认**值为 `./data` 保持兼容）
- 不修改上传文件存储目录结构（仍为 `data/uploads/<hash>`）

**验收标准**：
- 从任意目录 `go run cmd/server/main.go` 启动，上传功能正常
- `DATA_DIR=/tmp/testdata` 时上传文件落盘到 `/tmp/testdata/uploads`

---

### OPT-26 Magic Number 配置化

- **优先级**：P2
- **类别**：规范
- **风险等级**：低
- **影响文件**：`backend/internal/config/config.go`、`backend/internal/service/ranking.go`、`backend/internal/service/exam.go`

**现状**：
- `ranking.go:9-17` 含 `examScoreHalfLifeDay=30` 等 magic number
- 考试时长、AI 任务超时、邮件验证码有效期等同样硬编码

**变更要求**：
1. 在 `config.go` 增加配置项：
   - `RankingScoreHalfLifeDay = 30`
   - `ExamDefaultDurationMin = 120`
   - `AITaskTimeoutMin = 10`
   - `EmailCodeTTLMin = 10`
2. 各 service 通过构造函数注入 `*config.Config`
3. `.env.example` 增加对应变量

**禁止变更**：
- 不修改现有默认值（生产环境行为不变）
- 不修改接口请求/响应

**验收标准**：
- 启动后排行榜分数计算结果与优化前完全一致（默认值 30）
- 修改 `RANKING_SCORE_HALF_LIFE_DAY=60` 后排行榜分数按新值计算
- `go build ./...` 通过

---

### OPT-27 /healthz 健康检查端点

- **优先级**：P2
- **类别**：可观测性
- **风险等级**：低
- **影响文件**：`backend/internal/handler/health.go`（新建）、`backend/internal/router/router.go`

**现状**：
- 仅 `/ping` 端点，不检查 DB 连接
- K8s/PM2 探针不友好

**变更要求**：
1. 新建 `handler/health.go`：
   - `GET /healthz` 检查项：
     - DB ping（`db.Exec("SELECT 1")`）
     - 当前 goroutine 数量（`runtime.NumGoroutine()`）
     - 版本号（从 `const Version = "1.0.0"` 读取）
   - 全部正常返回 200 + `{"status": "ok", "db": "up", "goroutines": 50, "version": "1.0.0"}`
   - 任意失败返回 503
2. 在 router 中**不**挂载 `auth` 中间件（公开访问）

**禁止变更**：
- 不修改 `/ping` 行为（如有其他用途）
- 不修改现有业务接口

**验收标准**：
- DB 正常时 `curl /healthz` 返回 200
- DB 不可达时返回 503
- `go build ./...` 通过

---

### OPT-28 结构化日志

- **优先级**：P2
- **类别**：可观测性
- **风险等级**：中
- **影响文件**：`backend/cmd/server/main.go`、所有 `*.go` 文件中 `log.Printf` 调用

**现状**：
- 使用 Go 标准 `log` 包，输出非结构化
- 无法按 requestID 聚合日志

**变更要求**：
1. 引入 `go.uber.org/zap`
2. `main.go` 初始化 `zap.NewProduction()` 或 `NewDevelopment()`（根据 `APP_ENV`）
3. `gin.LoggerWithFormatter` 改为 zap adapter
4. 在中间件中注入 `requestID`（用 `uuid.NewString()`）到 context
5. 全局替换 `log.Printf` 为 `logger.Info(...).Str("request_id", id).Msg(...)`

**禁止变更**：
- 不修改现有日志输出的语义（如某个 `log.Printf("用户 %d 登录成功", uid)` 仍需能看出是登录成功）
- 不修改 HTTP 响应内容

**验收标准**：
- 每条日志包含 `request_id`、`timestamp`、`level`
- 同一次请求的所有日志 `request_id` 一致
- 现有 `log.Printf` 调用全部被替换（`grep -r "log.Printf" internal/` 仅为空或仅在新 logger 包装中）
- `go build ./...` 通过

---

### OPT-29 Dockerfile 与 docker-compose

- **优先级**：P2
- **类别**：部署
- **风险等级**：中
- **影响文件**：`backend/Dockerfile`（新建）、`docker-compose.yml`（项目根新建）

**现状**：
- 无容器化方案，部署靠手编 zip + scp
- CI/CD 无法自动化

**变更要求**：
1. `backend/Dockerfile`：
   - 多阶段构建：第一阶段 `golang:1.22-alpine` 编译，第二阶段 `alpine:3.19` 仅含可执行文件
   - 暴露 `8080` 端口
   - 非 root 用户运行
2. `docker-compose.yml`（项目根）：
   - 服务 `backend`：build `./backend`，挂载 `./data` 目录，依赖 `db`
   - 服务 `db`：`mysql:8.0`，挂载 `./data/mysql`，初始化执行 `database.sql` + `seed_*.sql`
3. `.dockerignore` 排除 `node_modules`、`dist`、`*.log`

**禁止变更**：
- 不修改应用代码
- 不修改 SQL 文件

**验收标准**：
- `docker-compose up` 一键启动后端 + 数据库
- `curl http://localhost:8080/ping` 返回 200
- 数据库 seed 数据可见
- 镜像构建成功，运行时无 root 进程

---

### OPT-30 优雅关闭（signal.NotifyContext）

- **优先级**：P2
- **类别**：可观测性 / 部署
- **风险等级**：中
- **影响文件**：`backend/cmd/server/main.go`

**现状**：
- `r.Run(":8080")` 是阻塞调用，SIGTERM 时进程被强杀
- K8s 滚动更新时旧 Pod 直接中断请求

**变更要求**：
1. 改用 `http.Server` + `srv.ListenAndServe()`
2. 监听 SIGTERM/SIGINT：触发 `srv.Shutdown(ctx)` 等待 30 秒
3. 期间所有 janitor goroutine 收到 cancel 信号（参考 OPT-20）
4. DB 连接优雅关闭

**禁止变更**：
- 不修改路由注册逻辑
- 不修改 HTTP handler 行为

**验收标准**：
- 启动后 `kill -TERM <pid>`，进程在 30 秒内退出
- 退出时日志包含 `shutting down` / `cleanup done`
- 关闭期间的新请求被拒绝（返回 503）

---

### OPT-31 软删除（gorm.DeletedAt）

- **优先级**：P2
- **类别**：可维护性
- **风险等级**：高（涉及 schema 与全量数据迁移）
- **影响文件**：`backend/internal/model/*.go`、`backend/internal/migrations/sql/`（新增）

**现状**：
- 24 张表**全部硬删除**，用户删除级联清空所有错题、考试、通知
- 误删无恢复途径

**变更要求**：
1. **评估**：与产品确认哪些表需要软删除（建议至少 `users` / `questions` / `exam_records` / `wrong_questions`）
2. 对需软删除的表：
   - 新增列 `deleted_at DATETIME NULL DEFAULT NULL`，加索引 `idx_deleted_at`
   - model 增加 `gorm.DeletedAt` 字段
   - GORM 自动 `WHERE deleted_at IS NULL`
3. 编写 migration 包含数据回滚（`down.sql` 中删除列）
4. 现有调用 `db.Delete(&user)` 改为 `db.Delete(&user)`（GORM 默认走 soft delete）；调用 `Unscoped().Delete(...)` 表示硬删

**禁止变更**：
- 不修改现有 API 行为（外部调用方无感）
- 不修改外键约束策略
- 不可一次性对全部 24 张表启用，先做 `users` 与 `questions` 两张试点

**验收标准**：
- `DELETE FROM users WHERE id=1` 后，`SELECT * FROM users WHERE id=1` 返回空，但 `SELECT * FROM users WHERE id=1 AND deleted_at IS NOT NULL` 能查到
- 用户关联的错题、考试记录仍可通过外键访问（不会因 CASCADE 丢失）
- 现有 API 行为不变
- `go build ./...` 与 migration 双向回滚测试通过

---

### OPT-32 DB 连接池自适应

- **优先级**：P2
- **类别**：性能
- **风险等级**：低
- **影响文件**：`backend/internal/database/database.go:78-81`

**现状**：
- `SetMaxOpenConns(20)`, `SetMaxIdleConns(10)` 写死
- AI 高并发任务时排队严重

**变更要求**：
1. 改为 `maxOpen = runtime.NumCPU() * 4`（下限 20，上限 100）
2. `maxIdle = maxOpen / 2`
3. `connMaxLifetime = 30 * time.Minute`
4. 启动日志打印实际配置

**禁止变更**：
- 不修改 DSN
- 不修改 `db.Exec` 等调用

**验收标准**：
- 8 核机器上 `db.Stats().MaxOpenConnections == 32`
- 高并发（100 QPS）下连接池无 `wait_count` 持续增长
- `go build ./...` 通过

---

### OPT-33 request 层错误信息优化

- **优先级**：P2
- **类别**：质量
- **风险等级**：低
- **影响文件**：`frontend/src/api/request.ts`

**现状**：
- 所有错误都 `showToast(json.message || '请求失败')`
- 401 重复 toast 跳转登录
- 业务预期错误（参数校验）也 toast

**变更要求**：
1. `request.ts` 增加 `silent: boolean` 选项（默认 false）
2. 业务预期错误（如重复收藏）调用方传 `silent: true`，request 不 toast
3. 401 跳转后**不再** toast
4. 网络错误（fetch reject）显示统一文案 `网络异常，请检查连接`

**禁止变更**：
- 不修改后端响应格式
- 不修改默认行为（多数接口仍正常 toast）

**验收标准**：
- 重复收藏（后端 409）不弹 toast
- 401 跳转到 `/login` 不弹任何 toast
- 断网时所有接口统一显示 `网络异常`
- `npm run typecheck` 通过

---

### OPT-34 AI 用户答案长度截断

- **优先级**：P2
- **类别**：安全 / 性能
- **风险等级**：低
- **影响文件**：`backend/internal/service/ai.go:1188-1198`

**现状**：
- `buildEssayScorePrompt` 中 `req.UserAnswer` 直接拼入 LLM prompt
- 用户答案无长度上限，**可能撑爆 token 或被 LLM 拒收**
- 超长答案还会**显著增加 API 成本**

**变更要求**：
1. 在 service 入口处增加：
   ```go
   const maxUserAnswerLen = 4000
   if len(req.UserAnswer) > maxUserAnswerLen {
       req.UserAnswer = req.UserAnswer[:maxUserAnswerLen] + "...(已截断)"
   }
   ```
2. 在 dto 中加注释说明截断位置

**禁止变更**：
- 不修改 LLM 调用逻辑
- 不修改响应字段
- 不修改 prompt 模板（仅截断 user_answer 部分）

**验收标准**：
- 用户提交 10000 字符的答案，实际传入 LLM 的为前 4000 字符 + 截断提示
- 评分结果与未截断时差异在可接受范围（<5%）
- `go build ./...` 通过

---

## 4. 总结

| 优先级 | 任务数 | 预估总工时 |
|--------|--------|-----------|
| P0 | 8 | 5-7 人日 |
| P1 | 14 | 10-15 人日 |
| P2 | 12 | 8-12 人日 |
| **合计** | **34** | **23-34 人日** |

### 建议执行顺序

1. **第一周**：OPT-01 → OPT-02 → OPT-03 → OPT-04 → OPT-05 → OPT-06 → OPT-07 → OPT-08（P0 全部）
2. **第二周**：OPT-13 → OPT-14 → OPT-15 → OPT-16（性能优化）；OPT-09 → OPT-10 → OPT-11 → OPT-12（安全）
3. **第三周**：OPT-17 → OPT-18 → OPT-19 → OPT-20 → OPT-21 → OPT-22（代码质量）
4. **第四周及之后**：OPT-23 ~ OPT-34 按需排期

### 风险提示

- **OPT-31（软删除）**风险最高，必须先与产品确认范围，并在测试环境充分验证
- **OPT-04（Admin 中间件）**部署前需在测试环境验证所有 admin 路由的访问控制
- **OPT-07（并发唯一约束）**migration 前必须先在测试环境模拟并发场景验证

---

**维护说明**：
- 任务完成后请将状态从 `pending` 改为 `completed`，并在任务卡片下追加 commit hash
- 任何发现需要新增的优化项，请使用 `OPT-NN` 递增编号，并按相同模板补充
