#!/usr/bin/env bash
# 软考学系后端启动脚本（Linux）
# 用法：
#   ./run.sh start      后台启动，日志输出到 logs/server.log
#   ./run.sh stop       停止
#   ./run.sh restart    重启
#   ./run.sh status     查看状态
#   ./run.sh foreground 前台运行（看实时日志，Ctrl+C 退出）
set -e

APP_NAME="softteststudyt-backend"
APP_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN="$APP_DIR/$APP_NAME"
PID_FILE="$APP_DIR/$APP_NAME.pid"
LOG_DIR="$APP_DIR/logs"
LOG_FILE="$LOG_DIR/server.log"

mkdir -p "$LOG_DIR"

is_running() {
  [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null
}

start_bg() {
  if is_running; then
    echo "[run] 已在运行，PID=$(cat "$PID_FILE")"
    return 0
  fi
  cd "$APP_DIR"
  nohup "$BIN" >> "$LOG_FILE" 2>&1 &
  echo $! > "$PID_FILE"
  sleep 1
  if is_running; then
    echo "[run] 已启动，PID=$(cat "$PID_FILE")，日志: $LOG_FILE"
  else
    echo "[run] 启动失败，请查看日志: $LOG_FILE"
    return 1
  fi
}

stop_app() {
  if ! is_running; then
    echo "[run] 未运行"
    rm -f "$PID_FILE"
    return 0
  fi
  PID="$(cat "$PID_FILE")"
  kill "$PID"
  for i in 1 2 3 4 5 6 7 8 9 10; do
    if ! kill -0 "$PID" 2>/dev/null; then break; fi
    sleep 0.5
  done
  if kill -0 "$PID" 2>/dev/null; then
    kill -9 "$PID" || true
  fi
  rm -f "$PID_FILE"
  echo "[run] 已停止"
}

case "${1:-start}" in
  start)   start_bg ;;
  stop)    stop_app ;;
  restart) stop_app; start_bg ;;
  status)
    if is_running; then echo "[run] 运行中，PID=$(cat "$PID_FILE")"
    else echo "[run] 未运行"; fi
    ;;
  foreground)
    cd "$APP_DIR"
    exec "$BIN"
    ;;
  *) echo "用法: $0 {start|stop|restart|status|foreground}"; exit 1 ;;
esac
