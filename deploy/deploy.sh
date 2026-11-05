#!/bin/sh
# 在 Ubuntu 服务器上以 root 执行： sudo sh deploy.sh
# 前置：同目录下放 lms-server（Linux amd64）、library_data.db（可选）、lms-secret.txt
set -e
cd "$(dirname "$0")"

echo "[1/5] 创建用户与目录"
id lms >/dev/null 2>&1 || useradd -r -s /sbin/nologin lms
mkdir -p /opt/lms

echo "[2/5] 复制文件"
cp -f lms-server /opt/lms/lms-server
chmod 755 /opt/lms/lms-server
if [ ! -f /opt/lms/library_data.db ]; then
  cp -f library_data.db /opt/lms/ 2>/dev/null || true
  [ -f library_data.db-wal ] && cp -f library_data.db-wal /opt/lms/ || true
  [ -f library_data.db-shm ] && cp -f library_data.db-shm /opt/lms/ || true
else
  echo "  /opt/lms/library_data.db 已存在，保留现有数据（不覆盖）"
fi
if [ ! -f lms-secret.txt ]; then
  tr -dc 'a-f0-9' < /dev/urandom | head -c 64 > lms-secret.txt
  echo "  未提供 lms-secret.txt，已随机生成"
fi
cp -f lms-secret.txt /opt/lms/lms-secret.txt
chmod 600 /opt/lms/lms-secret.txt
printf 'LMS_SECRET=%s\n' "$(tr -d '\r\n' < lms-secret.txt)" > /opt/lms/lms.env
chmod 600 /opt/lms/lms.env
chown -R lms:lms /opt/lms

echo "[3/5] 安装 systemd 服务"
cp -f lms-server.service /etc/systemd/system/lms-server.service
systemctl daemon-reload
systemctl enable lms-server >/dev/null
systemctl restart lms-server

echo "[4/5] 等待启动"
sleep 3

echo "[5/5] 健康检查"
curl -s --max-time 8 http://127.0.0.1:8080/api/health || echo "(无响应，检查 journalctl -u lms-server -n 50)"
echo
echo "完成。请确认：① 云安全组放行 TCP 8080；② Cloudflare SSL 模式 Flexible + Origin Rule 回源 8080。"
