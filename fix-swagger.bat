@echo off
REM Quick fix script for Swagger blank page issue (Windows)

echo.
echo ================================================================
echo   SWAGGER BLANK PAGE FIX
echo ================================================================
echo.
echo This will:
echo   1. Ensure Swagger docs are generated
echo   2. Rebuild backend Docker container
echo   3. Restart backend with Swagger included
echo.
set /p CONTINUE="Continue? (y/n): "
if /i not "%CONTINUE%"=="y" (
    echo.
    echo Cancelled.
    exit /b 1
)

echo.
echo ================================================================
echo   Step 1/4: Checking Swagger docs...
echo ================================================================
if exist "docs\docs.go" (
    echo [OK] Swagger docs found
) else (
    echo [WARN] Swagger docs not found, generating...
    make swagger
)

echo.
echo ================================================================
echo   Step 2/4: Stopping old backend container...
echo ================================================================
docker-compose stop backend

echo.
echo ================================================================
echo   Step 3/4: Rebuilding backend container...
echo   This may take 1-2 minutes...
echo ================================================================
docker-compose build --no-cache backend

echo.
echo ================================================================
echo   Step 4/4: Starting backend container...
echo ================================================================
docker-compose up -d backend

echo.
echo Waiting for backend to be healthy (15 seconds)...
timeout /t 15 /nobreak >nul

echo.
echo ================================================================
echo   Testing Swagger...
echo ================================================================
echo.

REM Test health endpoint
echo Testing health endpoint...
curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Health endpoint: Working
) else (
    echo [FAIL] Health endpoint: Not responding
)

REM Test Swagger JSON
echo Testing Swagger JSON endpoint...
curl -s http://localhost:8080/swagger/doc.json >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Swagger JSON: Available
) else (
    echo [FAIL] Swagger JSON: Not available
)

echo.
echo ================================================================
echo   FIX COMPLETE!
echo ================================================================
echo.
echo Open Swagger UI in your browser:
echo   http://localhost:8080/swagger/index.html
echo.
echo If still blank, try:
echo   1. Hard refresh: Ctrl+Shift+R
echo   2. Check logs: make docker-logs-backend
echo   3. Full reset: make docker-clean ^&^& make docker-setup
echo.
echo Press any key to open Swagger in browser...
pause >nul

start http://localhost:8080/swagger/index.html

echo.
echo Done!
echo.

