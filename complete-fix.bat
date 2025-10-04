@echo off
REM Complete fix for Swagger not running (Windows)

title FleetTracker - Complete Fix

echo.
echo ========================================================================
echo                    COMPLETE FIX - SWAGGER + BACKEND
echo ========================================================================
echo.
echo Current Issue: Swagger page shows "This site can't be reached"
echo.
echo This will:
echo   [1] Stop all containers
echo   [2] Remove problematic TimescaleDB data
echo   [3] Start everything fresh with fixed configuration
echo   [4] Wait for all services to be healthy (2 minutes)
echo   [5] Test Swagger and open in browser
echo.
echo ========================================================================
echo.
set /p CONTINUE="Ready to fix? This will delete TimescaleDB data. (y/n): "
if /i not "%CONTINUE%"=="y" (
    echo.
    echo Fix cancelled.
    pause
    exit /b 0
)

echo.
echo ========================================================================
echo   STEP 1/5: Stopping all containers...
echo ========================================================================
docker-compose down
if %errorlevel% neq 0 (
    echo [ERROR] Failed to stop containers
    pause
    exit /b 1
)
echo [OK] All containers stopped

echo.
echo ========================================================================
echo   STEP 2/5: Removing TimescaleDB volume...
echo ========================================================================
docker volume rm backend_timescale_data 2>nul
if %errorlevel% equ 0 (
    echo [OK] TimescaleDB volume removed
) else (
    echo [INFO] No existing TimescaleDB volume found (OK)
)

echo.
echo ========================================================================
echo   STEP 3/5: Starting all services...
echo ========================================================================
echo This may take a moment to download images...
echo.
docker-compose up -d
if %errorlevel% neq 0 (
    echo [ERROR] Failed to start services
    echo.
    echo Check logs with: docker-compose logs
    pause
    exit /b 1
)
echo [OK] Services started

echo.
echo ========================================================================
echo   STEP 4/5: Waiting for services to be healthy...
echo ========================================================================
echo.
echo This takes about 2 minutes. Please wait...
echo.
echo Status:
docker-compose ps
echo.

REM Wait in intervals and show progress
echo [00:30] Waiting... (PostgreSQL starting)
timeout /t 30 /nobreak >nul

echo [01:00] Waiting... (TimescaleDB initializing)
timeout /t 30 /nobreak >nul

echo [01:30] Waiting... (Backend starting)
timeout /t 30 /nobreak >nul

echo [02:00] Waiting... (Final checks)
timeout /t 30 /nobreak >nul

echo.
echo [OK] Wait complete

echo.
echo ========================================================================
echo   STEP 5/5: Testing services...
echo ========================================================================
echo.

REM Check container status
echo [TEST 1] Container Status:
docker-compose ps
echo.

REM Test backend health endpoint
echo [TEST 2] Backend Health:
curl -s http://localhost:8080/health
if %errorlevel% equ 0 (
    echo.
    echo [OK] Backend is responding
) else (
    echo.
    echo [WARN] Backend not responding yet, may need more time
)
echo.

REM Test Swagger JSON
echo [TEST 3] Swagger API Documentation:
curl -s http://localhost:8080/swagger/doc.json -o nul
if %errorlevel% equ 0 (
    echo [OK] Swagger API docs available
) else (
    echo [WARN] Swagger API docs not ready yet
)
echo.

REM Check logs for success messages
echo [TEST 4] Checking Backend Logs:
docker-compose logs backend | findstr /C:"API starting" >nul
if %errorlevel% equ 0 (
    echo [OK] Backend started successfully
) else (
    echo [WARN] Backend may still be starting
)
echo.

echo ========================================================================
echo                           FIX COMPLETE!
echo ========================================================================
echo.
echo Services Status:
docker-compose ps
echo.
echo ========================================================================
echo   NEXT STEPS:
echo ========================================================================
echo.
echo 1. Open Swagger UI: http://localhost:8080/swagger/index.html
echo 2. If blank page, wait 30 seconds and hard refresh (Ctrl+Shift+R)
echo 3. Check health: http://localhost:8080/health
echo.
echo Useful Commands:
echo   View logs:       docker-compose logs backend
echo   Restart:         docker-compose restart backend
echo   Stop all:        docker-compose down
echo.
echo ========================================================================
echo.
set /p OPEN="Open Swagger in browser now? (y/n): "
if /i "%OPEN%"=="y" (
    echo.
    echo Opening Swagger UI...
    start http://localhost:8080/swagger/index.html
    echo.
    echo If page is blank:
    echo   - Wait 30 seconds
    echo   - Press Ctrl+Shift+R to hard refresh
    echo   - Check logs: docker-compose logs backend
)

echo.
echo ========================================================================
echo.
echo If you still have issues:
echo   1. Check logs: docker-compose logs
echo   2. Read guide: COMPLETE_FIX.md
echo   3. Contact support
echo.
echo ========================================================================
echo.
pause

