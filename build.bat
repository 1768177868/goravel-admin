@echo off
chcp 65001 >nul
REM 构建脚本 - 支持指定不同的 .env 文件 (Windows)
REM 使用方法: build.bat [env_file] [output_name] [target_os] [target_arch]
REM 示例: 
REM   build.bat .env.production           REM 使用生产配置，生成 Linux 二进制文件 (main)
REM   build.bat .env.local main linux amd64  REM 指定所有参数
REM   build.bat .env.production main windows amd64  REM 生成 Windows 可执行文件
REM   build.bat                           REM 默认使用 .env，生成 Linux 二进制文件

setlocal enabledelayedexpansion

set "ENV_FILE=%~1"
if "%ENV_FILE%"=="" set "ENV_FILE=.env"

set "OUTPUT_NAME=%~2"
if "%OUTPUT_NAME%"=="" set "OUTPUT_NAME=main"

set "TARGET_OS=%~3"
if "%TARGET_OS%"=="" set "TARGET_OS=linux"

set "TARGET_ARCH=%~4"
if "%TARGET_ARCH%"=="" set "TARGET_ARCH=amd64"

echo ==========================================
echo Build Configuration
echo ==========================================
echo Environment File: %ENV_FILE%
echo Output File: %OUTPUT_NAME%
echo Target OS: %TARGET_OS%
echo Target Arch: %TARGET_ARCH%
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

REM 保存原始的 Go 环境变量
set "ORIGINAL_GOOS=%GOOS%"
set "ORIGINAL_GOARCH=%GOARCH%"
set "ORIGINAL_CGO_ENABLED=%CGO_ENABLED%"

REM 设置跨平台编译环境变量
set GOOS=%TARGET_OS%
set GOARCH=%TARGET_ARCH%
set CGO_ENABLED=0

REM 执行构建（Linux 二进制文件）
go build --ldflags "-extldflags -static" -o "%OUTPUT_NAME%" .

set BUILD_RESULT=%errorlevel%

REM 恢复原始的 Go 环境变量
if defined ORIGINAL_GOOS (
    set GOOS=%ORIGINAL_GOOS%
) else (
    set GOOS=
)
if defined ORIGINAL_GOARCH (
    set GOARCH=%ORIGINAL_GOARCH%
) else (
    set GOARCH=
)
if defined ORIGINAL_CGO_ENABLED (
    set CGO_ENABLED=%ORIGINAL_CGO_ENABLED%
) else (
    set CGO_ENABLED=
)

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

