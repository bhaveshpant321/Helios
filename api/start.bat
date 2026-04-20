@echo off
REM Helios API Quick Start Script for Windows
REM This script helps you set up and run the Helios API server

echo ========================================
echo Helios API Quick Start
echo ========================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed. Please install Go 1.21 or later.
    pause
    exit /b 1
)

echo [OK] Go is installed
go version
echo.

REM Check if .env exists
if not exist ".env" (
    echo [WARNING] .env file not found. Creating from .env.example...
    copy .env.example .env >nul
    echo [INFO] Please edit .env file with your database credentials
    echo        Then run this script again.
    pause
    exit /b 0
)

echo [OK] Found .env configuration
echo.

REM Check if go.mod exists
if not exist "go.mod" (
    echo [ERROR] go.mod not found. Please run this script from the api directory.
    pause
    exit /b 1
)

REM Download dependencies
echo [INFO] Downloading Go dependencies...
go mod download
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to download dependencies
    pause
    exit /b 1
)
echo [OK] Dependencies downloaded
echo.

REM Build the application
echo [INFO] Building application...
go build -o helios-api.exe main.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Build failed
    pause
    exit /b 1
)
echo [OK] Build successful
echo.

REM Run the application
echo [INFO] Starting Helios API server...
echo        Press Ctrl+C to stop the server
echo.
helios-api.exe
