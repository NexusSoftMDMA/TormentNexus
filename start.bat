@echo off
setlocal EnableDelayedExpansion

for /f "tokens=*" %%v in ('type VERSION') do set VER=%%v

echo ==============================================
echo  Borg AIOS v%VER%
echo  The Neural Operating System
echo ================================================
echo.

REM -- 1. Build Go Sidecar --------------------------
where go >nul 2>nul
if errorlevel 1 (
    echo [SKIP] Go not found - skipping Go sidecar build.
    goto :skip_go
)
echo [1/5] Building Go sidecar...
cd go
go build -ldflags "-X github.com/borghq/borg-go/internal/buildinfo.Version=%VER%" -buildvcs=false -o ..\bin\borg.exe ./cmd/borg 2>nul
if errorlevel 1 (
    echo [WARN] Go build failed - continuing without sidecar.
) else (
    echo   OK bin/borg.exe built
)
cd ..
:skip_go

REM -- 2. Run Readiness Probe ------------------------
set SKIP_READINESS=0
if /I "%BORG_SKIP_READINESS%"=="1" set SKIP_READINESS=1
if "%SKIP_READINESS%"=="1" (
    echo [2/5] Skipping readiness probe (BORG_SKIP_READINESS=1)
) else (
    echo [2/5] Checking dev readiness...
    node scripts/verify_dev_readiness.mjs --soft
)

REM -- 3. Install Dependencies ----------------------
where pnpm >nul 2>nul
if errorlevel 1 (
    echo [FATAL] pnpm not found. Install with: npm install -g pnpm
    exit /b 1
)
set SKIP_INSTALL=0
if /I "%BORG_SKIP_INSTALL%"=="1" set SKIP_INSTALL=1
if "%SKIP_INSTALL%"=="1" (
    echo [3/5] Skipping install (BORG_SKIP_INSTALL=1)
) else (
    echo [3/5] Installing dependencies...
    call pnpm install --frozen-lockfile 2>nul || call pnpm install
    if errorlevel 1 exit /b 1
)

REM -- 4. Build TypeScript ---------------------------
set SKIP_BUILD=0
if /I "%BORG_SKIP_BUILD%"=="1" set SKIP_BUILD=1
if "%SKIP_BUILD%"=="1" (
    echo [4/5] Skipping build (BORG_SKIP_BUILD=1)
) else (
    echo [4/5] Building TypeScript...
    call pnpm run build:workspace 2>nul || call pnpm -C packages/core exec tsc && pnpm -C packages/cli exec tsc
    if errorlevel 1 (
        echo [WARN] Build had issues - continuing anyway.
    )
)

REM -- 5. Launch Services ---------------------------
echo [5/5] Starting services...
echo.

REM Start Go sidecar in background if binary exists
if exist bin\borg.exe (
    curl -s -o nul -w "%%{http_code}" http://127.0.0.1:4300/health > "%TEMP%\borg_go_check.txt" 2>nul
    set /p GO_CHECK=<"%TEMP%\borg_go_check.txt"
    if "!GO_CHECK!"=="200" (
        echo   Go sidecar already running on port 4300.
    ) else (
        echo   Starting Go sidecar on port 4300...
        start /B bin\borg.exe -port 4300 > nul 2>&1
    )
)

REM Start Next.js dashboard in background
set DASH_PORT=%BORG_DASH_PORT%
if "%DASH_PORT%"=="" set DASH_PORT=3000
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%DASH_PORT%/dashboard > "%TEMP%\borg_web_check.txt" 2>nul
set /p WEB_CHECK=<"%TEMP%\borg_web_check.txt"
if "!WEB_CHECK!"=="200" (
    echo   Next.js dashboard already running on port %DASH_PORT%.
) else (
    echo   Starting Next.js dashboard on port %DASH_PORT%...
    start /B cmd /c "cd apps\web && npx next dev --port %DASH_PORT% > nul 2>&1"
)

REM Start TypeScript control plane
set BORG_PORT=%BORG_PORT%
if "%BORG_PORT%"=="" set BORG_PORT=4100

REM Check if core is already running
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%BORG_PORT%/health > "%TEMP%\borg_core_check.txt" 2>nul
set /p CORE_CHECK=<"%TEMP%\borg_core_check.txt"
if "!CORE_CHECK!"=="200" (
    echo   TS control plane already running on port %BORG_PORT%.
) else (
    echo   Starting TS control plane on port %BORG_PORT%...
)
echo.
echo   tRPC:      http://0.0.0.0:%BORG_PORT%/trpc
echo   REST:      http://0.0.0.0:%BORG_PORT%/api
echo   Go:        http://127.0.0.1:4300/health
echo   Dashboard: http://localhost:%DASH_PORT%/dashboard
echo.
if not "!CORE_CHECK!"=="200" (
    echo Press Ctrl+C to stop all services.
    echo.
    node packages/cli/dist/cli/src/index.js start --port %BORG_PORT% %*
) else (
    echo Press Ctrl+C to stop all services.
    echo.
    echo All services running. Core was already active.
    timeout /t 99999 > nul
)
