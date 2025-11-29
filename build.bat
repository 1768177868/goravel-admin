@echo off
chcp 65001 >nul
REM 构建脚本 - 支持指定不同的 .env 文件 (Windows)
REM 使用方法: build.bat [env_file] [output_name]
REM 示例: 
REM   build.bat .env.local      # 使用本地配置
REM   build.bat .env.production # 使用生产配置
REM   build.bat                 # 默认使用 .env

setlocal enabledelayedexpansion

set "ENV_FILE=%~1"
if "%ENV_FILE%"=="" set "ENV_FILE=.env"

set "OUTPUT_NAME=%~2"
if "%OUTPUT_NAME%"=="" set "OUTPUT_NAME=main.exe"

echo ==========================================
echo Build Configuration
echo ==========================================
echo Environment File: %ENV_FILE%
echo Output File: %OUTPUT_NAME%
echo ==========================================

REM 检查指定的环境文件是否存在
if not exist "%ENV_FILE%" (
    echo Error: Environment file %ENV_FILE% does not exist!
    echo Please create %ENV_FILE% file first
    exit /b 1
)

REM 如果存在 .env 文件，先备份为 .env.bak
if exist ".env" (
    copy /Y .env .env.bak >nul 2>&1
    echo Backed up .env to .env.bak
)

REM 复制指定的环境文件为 .env（构建时使用）
copy /Y "%ENV_FILE%" .env >nul 2>&1

echo Copied %ENV_FILE% to .env
echo Starting build...

REM 执行构建
go build --ldflags "-extldflags -static" -o "%OUTPUT_NAME%" .

set BUILD_RESULT=%errorlevel%

REM 构建完成后，恢复原来的 .env 文件
if exist ".env.bak" (
    move /Y .env.bak .env >nul 2>&1
    echo Restored .env.bak to .env
)

if %BUILD_RESULT% equ 0 (
    echo ==========================================
    echo Build successful!
    echo Output file: %OUTPUT_NAME%
    echo ==========================================
    exit /b 0
) else (
    echo ==========================================
    echo Build failed!
    echo ==========================================
    exit /b 1
)

