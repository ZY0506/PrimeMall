@echo off
chcp 65001 >nul
title PrimeMall 一键停止
cd /d "%~dp0"

echo ============================================
echo   PrimeMall - 停止所有服务
echo ============================================
echo.

echo [1/3] 停止 Nginx...
docker compose stop nginx-shop 2>nul
echo   ✅ Nginx 已停止
echo.

echo [2/3] 停止网关实例...
taskkill /F /IM shop.exe 2>nul
echo   ✅ 3 个网关实例已停止
echo.

echo [3/3] 停止 RPC 后端服务...
taskkill /F /IM user.exe 2>nul
taskkill /F /IM product.exe 2>nul
taskkill /F /IM order.exe 2>nul
taskkill /F /IM payment.exe 2>nul
taskkill /F /IM marketing.exe 2>nul
taskkill /F /IM admin.exe 2>nul
taskkill /F /IM search.exe 2>nul
taskkill /F /IM admin-api.exe 2>nul
echo   ✅ RPC 后端服务已停止

echo.
echo   ⚠️  Docker 基础设施未停止（MySQL/Redis等继续运行）
echo      如需完全停止，执行: docker compose down
echo.
echo   ✅ PrimeMall 已停止
echo.
pause
