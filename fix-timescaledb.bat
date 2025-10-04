@echo off
REM Quick fix script for TimescaleDB startup issues (Windows)

echo.
echo ================================================================
echo   TIMESCALEDB STARTUP FIX
echo ================================================================
echo.
echo This will:
echo   1. Stop all containers
echo   2. Remove old TimescaleDB data
echo   3. Rebuild with simplified init script
echo   4. Start everything fresh
echo.
echo WARNING: This will DELETE all TimescaleDB data!
echo.
set /p CONTINUE="Continue? (y/n): "
if /i not "%CONTINUE%"=="y" (
    echo.
    echo Cancelled.
    exit /b 1
)

echo.
echo ================================================================
echo   Step 1/5: Stopping all containers...
echo ================================================================
docker-compose down

echo.
echo ================================================================
echo   Step 2/5: Removing TimescaleDB volume...
echo ================================================================
docker volume rm backend_timescale_data 2>nul
if %errorlevel% equ 0 (
    echo [OK] TimescaleDB volume removed
) else (
    echo [INFO] No existing volume to remove
)

echo.
echo ================================================================
echo   Step 3/5: Checking init script...
echo ================================================================
if exist "init-timescale.sql" (
    echo [OK] Init script found
) else (
    echo [ERROR] init-timescale.sql not found!
    exit /b 1
)

echo.
echo ================================================================
echo   Step 4/5: Starting services...
echo ================================================================
docker-compose up -d

echo.
echo ================================================================
echo   Step 5/5: Waiting for services (30 seconds)...
echo ================================================================
timeout /t 30 /nobreak >nul

echo.
echo ================================================================
echo   Checking service status...
echo ================================================================
docker-compose ps

echo.
echo ================================================================
echo   Checking TimescaleDB logs...
echo ================================================================
docker-compose logs timescaledb --tail 20

echo.
echo ================================================================
echo   FIX COMPLETE!
echo ================================================================
echo.
echo Check if TimescaleDB is healthy:
echo   docker-compose ps
echo.
echo View full logs:
echo   make docker-logs-timescale
echo.
echo If still failing, check logs for errors:
echo   docker-compose logs timescaledb
echo.
pause

