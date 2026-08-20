# 软考学系后端部署说明（Linux amd64）

## 1. 目录结构

```
deploy/
├── softteststudyt-backend    # 后端可执行文件（已静态编译，glibc 无关）
├── .env                      # 环境变量配置（DB / JWT / 端口等）
├── run.sh                    # 启停脚本
├── data/
│   └── uploads/              # 学习资料上传文件目录（运行时由后端 /uploads 暴露）
└── logs/                     # 启动后自动生成，存放 server.log
```

## 2. 上传到服务器

将整个 `deploy/` 目录上传到服务器任意位置，例如：

```bash
scp -r deploy/ root@43.157.65.164:/opt/softteststudyt/
```

## 3. 首次部署

```bash
ssh root@43.157.65.164
cd /opt/softteststudyt
chmod +x run.sh softteststudyt-backend
./run.sh start
./run.sh status
tail -f logs/server.log
```

启动成功后会看到：

```
[config] 已加载环境变量文件: /opt/softteststudyt/.env
[config] DB=SoftTestStudyT@43.157.65.164:3306/softteststudyt  ServerPort=3000
...
[GIN-debug] Listening and serving HTTP on :3000
```

## 4. 常用命令

| 命令 | 说明 |
|------|------|
| `./run.sh start` | 后台启动 |
| `./run.sh stop` | 停止 |
| `./run.sh restart` | 重启 |
| `./run.sh status` | 查看状态 |
| `./run.sh foreground` | 前台运行（看实时日志） |
| `tail -f logs/server.log` | 跟踪日志 |

## 5. 防火墙 / 安全组

确保服务器 `3000` 端口已放行（云服务商安全组 + 系统防火墙 `firewall-cmd` / `ufw` / `iptables`）。

## 6. 修改配置

直接编辑 `.env` 后执行 `./run.sh restart` 生效。
如果不想用 `.env`，也可以用真实环境变量（`run.sh` 启动时 `nohup` 会继承当前 shell 的环境变量，优先级高于 `.env`）。

## 7. 开机自启（systemd，可选）

创建 `/etc/systemd/system/softteststudyt.service`：

```ini
[Unit]
Description=Soft Test Study Backend
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/softteststudyt
ExecStart=/opt/softteststudyt/softteststudyt-backend
Restart=always
RestartSec=3
EnvironmentFile=/opt/softteststudyt/.env

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable --now softteststudyt
systemctl status softteststudyt
```

> 用了 systemd 之后就用 `systemctl` 启停，不要再用 `run.sh` 避免双进程。
