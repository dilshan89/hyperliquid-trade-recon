@echo off
setlocal enabledelayedexpansion

echo =========================================
echo Building Hyperliquid Trade Reconciliation
echo =========================================
echo.

REM Step 1: Build React frontend
echo Step 1/3: Building React frontend...
cd frontend
call npm run build
if !errorlevel! neq 0 (
    echo Error: Frontend build failed
    exit /b 1
)
cd ..
echo [32m✓ Frontend build complete[0m
echo.

REM Step 2: Copy build to backend for embedding
echo Step 2/3: Preparing files for embedding...
if exist backend\frontend\build rmdir /s /q backend\frontend\build
if not exist backend\frontend mkdir backend\frontend
xcopy /E /I /Y frontend\build backend\frontend\build > nul
echo [32m✓ Files prepared for embedding[0m
echo.

REM Step 3: Build Go binary with embedded frontend
echo Step 3/3: Building Go binary with embedded frontend...
cd backend
set CGO_ENABLED=0
go build -o ..\hyperliquid-recon.exe main.go
if !errorlevel! neq 0 (
    echo Error: Go build failed
    cd ..
    exit /b 1
)
cd ..
echo [32m✓ Binary build complete[0m
echo.

REM Cleanup
echo Cleaning up...
rmdir /s /q backend\frontend
echo [32m✓ Cleanup complete[0m
echo.

echo =========================================
echo [32mBuild successful![0m
echo =========================================
echo.
echo Single executable created: hyperliquid-recon.exe
echo.
echo To run the application:
echo   hyperliquid-recon.exe
echo.
echo Then open your browser at: http://localhost:8080
echo.
