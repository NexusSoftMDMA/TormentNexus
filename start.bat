@echo off
setlocal EnableDelayedExpansion

for /f "tokens=*" %%v in ('type VERSION') do set VER=%%v

echo ================================================
echo  TormentNexus TORMENTNEXUS v%VER%
echo  The Neural Operating System (Go-Native)
echo ================================================
echo.

REM -- 1. Locate Go sidecar binary --
if not exist bin\tormentnexus.exe (
    echo [FATAL] Go sidecar binary not found. Please run build.bat first.
    exit /b 1
)

REM -- 2. Quick install dependencies (optional) --
set SKIP_INSTALL=0
if /I "%TORMENTNEXUS_SKIP_INSTALL%"=="1" set SKIP_INSTALL=1
if "%SKIP_INSTALL%"=="1" (
    echo   Skipping pnpm install (TORMENTNEXUS_SKIP_INSTALL=1)
) else (
    echo   Installing dependencies...
    call pnpm install --ignore-scripts 2>nul || call pnpm install
)

REM -- 3. Start Go Sidecar on port 4300 --
set GO_PORT=4300
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%GO_PORT%/health > "%TEMP%\tormentnexus_go_check.txt" 2>nul
set /p GO_CHECK=<"%TEMP%\tormentnexus_go_check.txt"
if "!GO_CHECK!"=="200" (
    echo   Go sidecar already running on port %GO_PORT%.
) else (
    echo   Starting Go sidecar on port %GO_PORT%...
    start /B bin\tormentnexus.exe serve --port %GO_PORT% > nul 2>&1
)

REM -- 4. Start Next.js dashboard on port 3000 --
set DASH_PORT=3000
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%DASH_PORT%/dashboard > "%TEMP%\tormentnexus_web_check.txt" 2>nul
set /p WEB_CHECK=<"%TEMP%\tormentnexus_web_check.txt"
if "!WEB_CHECK!"=="200" (
    echo   Next.js dashboard already running on port %DASH_PORT%.
) else (
    echo   Starting Next.js dashboard on port %DASH_PORT%...
    start /B cmd /c "cd apps\web && set NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1 && npx next dev --port %DASH_PORT% > nul 2>&1"
)

echo.
echo   Go Sidecar API: http://127.0.0.1:%GO_PORT%
echo   Go Health check: http://127.0.0.1:%GO_PORT%/health
echo   Dashboard UI:    http://localhost:%DASH_PORT%/dashboard
echo.
echo Press Ctrl+C to stop all services.
echo.
timeout /t 99999 > nul
