@echo off
chcp 65001 >nul
title PrimeMall 一键启动
cd /d "%~dp0"

echo ============================================
echo   PrimeMall - 一键启动所有服务
echo ============================================
echo.

REM ============================================
REM 第 1 步：构建所有服务
REM ============================================
echo [1/5] 构建服务二进制...
echo.
go build -o bin\user.exe apps\service\user\rpc\user.go
if %errorlevel% neq 0 ( echo ❌ user.rpc 构建失败 & pause & exit /b 1 )
echo   ✅ user.rpc

go build -o bin\product.exe apps\service\product\rpc\product.go
if %errorlevel% neq 0 ( echo ❌ product.rpc 构建失败 & pause & exit /b 1 )
echo   ✅ product.rpc

go build -o bin\order.exe apps\service\order\rpc\order.go
if %errorlevel% neq 0 ( echo ❌ order.rpc 构建失败 & pause & exit /b 1 )
echo   ✅ order.rpc

go build -o bin\payment.exe apps\service\payment\rpc\payment.go
if %errorlevel% neq 0 ( echo ❌ payment.rpc 构建失败 & pause & exit /b 1 )
echo   ✅ payment.rpc

go build -o bin\marketing.exe apps\service\marketing\rpc\marketing.go
if %errorlevel% neq 0 ( echo ❌ marketing.rpc 构建失败 & pause & exit /b 1 )
echo   ✅ marketing.rpc

go build -o bin\admin.exe apps\service\admin\rpc\admin.go
if %errorlevel% neq 0 ( echo ❌ admin.rpc 构建失败 & pause & exit /b 1 )
echo   ✅ admin.rpc

go build -o bin\search.exe apps\service\search\rpc\search.go
if %errorlevel% neq 0 ( echo ❌ search.rpc 构建失败 & pause & exit /b 1 )
echo   ✅ search.rpc

go build -o bin\shop.exe apps\gateway\shop\shop.go
if %errorlevel% neq 0 ( echo ❌ shop-api 构建失败 & pause & exit /b 1 )
echo   ✅ shop-api

go build -o bin\admin-api.exe apps\gateway\admin\admin.go
if %errorlevel% neq 0 ( echo ❌ admin-api 构建失败 & pause & exit /b 1 )
echo   ✅ admin-api

echo.
echo   ✅ 所有服务构建完成
echo.

REM ============================================
REM 第 2 步：启动 Docker 基础设施
REM ============================================
echo [2/5] 启动 Docker 基础设施...
docker compose up -d mysql redis etcd rabbitmq elasticsearch 2>nul
echo   ✅ MySQL / Redis / Etcd / RabbitMQ / ES
echo.

REM ============================================
REM 第 3 步：启动 RPC 后端服务
REM ============================================
echo [3/5] 启动 RPC 后端服务...

REM 先停旧的
taskkill /F /IM user.exe 2>nul
taskkill /F /IM product.exe 2>nul
taskkill /F /IM order.exe 2>nul
taskkill /F /IM payment.exe 2>nul
taskkill /F /IM marketing.exe 2>nul
taskkill /F /IM admin.exe 2>nul
taskkill /F /IM search.exe 2>nul
timeout /t 2 /nobreak >nul

REM 按依赖顺序启动（user / product 先启动，其他依赖它们）
start "user.rpc"      bin\user.exe -f apps\service\user\rpc\etc\user.yaml
timeout /t 2 /nobreak >nul

start "product.rpc"   bin\product.exe -f apps\service\product\rpc\etc\product.yaml
timeout /t 2 /nobreak >nul

start "payment.rpc"   bin\payment.exe -f apps\service\payment\rpc\etc\payment.yaml
start "marketing.rpc" bin\marketing.exe -f apps\service\marketing\rpc\etc\marketing.yaml
timeout /t 2 /nobreak >nul
start "order.rpc"     bin\order.exe -f apps\service\order\rpc\etc\order.yaml
start "search.rpc"    bin\search.exe -f apps\service\search\rpc\etc\search.yaml
start "admin.rpc"     bin\admin.exe -f apps\service\admin\rpc\etc\admin.yaml
timeout /t 3 /nobreak >nul
echo   ✅ 7 个 RPC 服务已启动
echo.

REM ============================================
REM 第 4 步：启动 3 个网关实例
REM ============================================
echo [4/5] 启动网关实例（负载均衡集群）...

taskkill /F /IM shop.exe 2>nul
timeout /t 1 /nobreak >nul

start "shop-9091"     bin\shop.exe -f apps\gateway\shop\etc\shop-api-9091.yaml
start "shop-9092"     bin\shop.exe -f apps\gateway\shop\etc\shop-api-9092.yaml
start "shop-9093"     bin\shop.exe -f apps\gateway\shop\etc\shop-api-9093.yaml
timeout /t 3 /nobreak >nul
echo   ✅ 3 个网关实例已启动 (9091/9092/9093)
echo.

REM ============================================
REM 第 5 步：启动 Nginx 负载均衡
REM ============================================
echo [5/5] 启动 Nginx 负载均衡...
docker compose up -d nginx-shop 2>nul
timeout /t 2 /nobreak >nul
echo   ✅ Nginx 已启动 (端口 8080)
echo.

REM ============================================
REM 验证
REM ============================================
echo ============================================
echo   验证服务状态
echo ============================================
echo.

echo   ┌─────────────────────┬──────────┐
for %%s in (9091 9092 9093) do (
  curl -s http://127.0.0.1:%%s/health >nul 2>&1 && (
    echo   │  shop-%%s              Running  │
  ) || (
    echo   │  shop-%%s              FAILED   │
  )
)
curl -s http://127.0.0.1:8080/health >nul 2>&1 && (
  echo   │  nginx:8080            Running  │
) || (
  echo   │  nginx:8080            FAILED   │
)
echo   └─────────────────────┴──────────┘

echo.
echo ============================================
echo   ✅ PrimeMall 启动完成!
echo   访问地址: http://127.0.0.1:8080
echo ============================================
echo.
pause
