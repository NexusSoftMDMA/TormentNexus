@echo off
setlocal EnableDelayedExpansion

REM ==============================================
REM  TormentNexus TORMENTNEXUS - TypeScript Core + Dashboard
REM  Starts the TS control plane and Next.js web UI
REM ==============================================

REM -- Configuration ----------------------------------
set TS_PORT=%TORMENTNEXUS_TS_PORT%
if "%TS_PORT%"=="" set TS_PORT=4100
set DASH_PORT=%TORMENTNEXUS_DASH_PORT%
if "%DASH_PORT%"=="" set DASH_PORT=3000
set TS_DIR=%TORMENTNEXUS_TS_DIR%
if "%TS_DIR%"=="" set TS_DIR=archive\ts-legacy

REM -- 1. Check Node/pnpm ----------------------------
where node >nul 2>nul
if errorlevel 1 (
    echo [FATAL] Node.js not found. Install from https://nodejs.org/
    exit /b 1
)
where pnpm >nul 2>nul
if errorlevel 1 (
    echo [1/5] Installing pnpm...
    call npm install -g pnpm
    if errorlevel 1 exit /b 1
) else (
    echo [1/5] pnpm found.
)

REM -- 2. Install Dependencies -----------------------
set SKIP_INSTALL=0
if /I "%TORMENTNEXUS_SKIP_INSTALL%"=="1" set SKIP_INSTALL=1
if "%SKIP_INSTALL%"=="1" (
    echo [2/5] Skipping install (TORMENTNEXUS_SKIP_INSTALL=1)
) else (
    echo [2/5] Installing dependencies...
    cd %TS_DIR%
    call pnpm install --frozen-lockfile 2>nul || call pnpm install
    if errorlevel 1 (
        echo [FATAL] pnpm install failed.
        cd ..\..
        exit /b 1
    )
    cd ..\..
)

REM -- 3. Build TypeScript ---------------------------
set SKIP_BUILD=0
if /I "%TORMENTNEXUS_SKIP_BUILD%"=="1" set SKIP_BUILD=1
if "%SKIP_BUILD%"=="1" (
    echo [3/5] Skipping build (TORMENTNEXUS_SKIP_BUILD=1)
) else (
    echo [3/5] Building TypeScript packages...
    cd %TS_DIR%
    call pnpm run build:workspace 2>nul
    if errorlevel 1 (
        echo [WARN] Build had issues - continuing anyway.
    )
    cd ..\..
)

REM -- 4. Start Next.js Dashboard --------------------
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%DASH_PORT%/dashboard > "%TEMP%\hc_dash_check.txt" 2>nul
set /p DASH_CHECK=<"%TEMP%\hc_dash_check.txt"
if "!DASH_CHECK!"=="200" (
    echo [4/5] Dashboard already running on port %DASH_PORT%.
) else (
    echo [4/5] Starting Next.js dashboard on port %DASH_PORT%...
    set NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1
    start /B cmd /c "cd %TS_DIR%\apps\web && set NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1 && npx next dev --port %DASH_PORT% > nul 2>&1"
    echo   Waiting for dashboard to compile...
    set WAIT=0
    :dash_wait
    if !WAIT! GEQ 30 goto :dash_done
    curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%DASH_PORT%/dashboard > "%TEMP%\hc_dash_wait.txt" 2>nul
    set /p DASH_READY=<"%TEMP%\hc_dash_wait.txt"
    if "!DASH_READY!"=="200" goto :dash_done
    set /a WAIT+=1
    timeout /t 2 > nul
    goto :dash_wait
    :dash_done
)

REM -- 5. Start TS Control Plane ---------------------
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%TS_PORT%/health > "%TEMP%\hc_ts_check.txt" 2>nul
set /p TS_CHECK=<"%TEMP%\hc_ts_check.txt"
if "!TS_CHECK!"=="200" (
    echo [5/5] TS control plane already running on port %TS_PORT%.
    echo.
    echo   tRPC:      http://127.0.0.1:%TS_PORT%/trpc
    echo   REST:      http://127.0.0.1:%TS_PORT%/api
    echo   Health:    http://127.0.0.1:%TS_PORT%/health
    echo   Dashboard: http://localhost:%DASH_PORT%/dashboard
    echo.
    echo Press Ctrl+C to exit.
    timeout /t 99999 > nul
) else (
    echo [5/5] Starting TS control plane on port %TS_PORT%...
    echo.
    echo   tRPC:      http://127.0.0.1:%TS_PORT%/trpc
    echo   REST:      http://127.0.0.1:%TS_PORT%/api
    echo   Health:    http://127.0.0.1:%TS_PORT%/health
    echo   Dashboard: http://localhost:%DASH_PORT%/dashboard
    echo.
    echo Press Ctrl+C to stop all services.
    echo.
    cd %TS_DIR%
    node packages\cli\dist\cli\src\index.js start --port %TS_PORT% %*
    cd ..\..
)

exit /b %errorlevel%
