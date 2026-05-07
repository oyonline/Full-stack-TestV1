#!/bin/bash
set -e
echo "build + migrate via Makefile"
go mod tidy
make build-and-migrate
chmod +x ./go-admin
echo "kill go-admin service"
killall go-admin || true # kill go-admin service
nohup ./go-admin server -c=config/settings.dev.yml >> access.log 2>&1 & #后台启动服务将日志写入access.log文件
echo "run go-admin success"
ps -aux | grep go-admin
