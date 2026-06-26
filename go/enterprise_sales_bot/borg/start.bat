@echo off
setlocal EnableDelayedExpansion

for /f "tokens=*" %%v in ('type VERSION') do set VER=%%v

REM --- Parse --runtime argument ---
set RUNTIME_MODE=auto
set REMAINING_ARGS=
:parse_args
if "%1"=="" goto :args_done
if /I "%1"=="--runtime" (
    set RUNTIME_MODE=%2
    shift
    shift
    goto :parse_args
)
set REMAINING_ARGS=!REMAINING_ARGS! %1
shift
goto :parse_args
:args_done

echo ==============================================
echo  TormentNexus TORMENTNEXUS v%VER%
echo  The Neural Operating System
echo ================================================
echo.

REM --- Detect Go binary ---
set HAS_GO_BIN=0
if exist tormentnexus.exe set HAS_GO_BIN=1
if exist bin\tormentnexus.exe set HAS_GO_BIN=1

REM --- Resolve runtime mode ---
if /I "!RUNTIME_MODE!"=="auto" (
    if "!HAS_GO_BIN!"=="1" (
        set RUNTIME_MODE=go
    ) else (
        set RUNTIME_MODE=node
    )
)
if /I "!RUNTIME_MODE!"=="go" (
    echo [RUNTIME] Go-primary mode
) else (
    echo [RUNTIME] TS-primary mode (Node.js)
)
echo.

if /I "!RUNTIME_MODE!"=="go" goto :go_primary

REM ===== TS-PRIMARY PATH (legacy) =====

REM -- 1. Build Go Sidecar --------------------------
where go >nul 2>nul
if errorlevel 1 (
    echo [SKIP] Go not found - skipping Go sidecar build.
    goto :skip_go_ts
)
echo [1/5] Building Go sidecar...
cd go
go build -ldflags "-s -w -X github.com/tormentnexushq/tormentnexus-go/internal/buildinfo.Version=%VER%" -buildvcs=false -o ..\bin\tormentnexus.exe ./cmd/tormentnexus 2>nul
if errorlevel 1 (
    echo [WARN] Go build failed - continuing without sidecar.
) else (
    echo   OK bin/tormentnexus.exe built
)
cd ..
:skip_go_ts

REM -- 2. Run Readiness Probe ------------------------
set SKIP_READINESS=0
if /I "%TORMENTNEXUS_SKIP_READINESS%"=="1" set SKIP_READINESS=1
if "%SKIP_READINESS%"=="1" (
    echo [2/5] Skipping readiness probe (TORMENTNEXUS_SKIP_READINESS=1)
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
if /I "%TORMENTNEXUS_SKIP_INSTALL%"=="1" set SKIP_INSTALL=1
if "%SKIP_INSTALL%"=="1" (
    echo [3/5] Skipping install (TORMENTNEXUS_SKIP_INSTALL=1)
) else (
    echo [3/5] Installing dependencies...
    call pnpm install --frozen-lockfile 2>nul || call pnpm install
    if errorlevel 1 exit /b 1
)

REM -- 4. Build TypeScript ---------------------------
set SKIP_BUILD=0
if /I "%TORMENTNEXUS_SKIP_BUILD%"=="1" set SKIP_BUILD=1
if "%SKIP_BUILD%"=="1" (
    echo [4/5] Skipping build (TORMENTNEXUS_SKIP_BUILD=1)
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
if exist bin\tormentnexus.exe (
    curl -s -o nul -w "%%{http_code}" http://127.0.0.1:7778/health > "%TEMP%\tormentnexus_go_check.txt" 2>nul
    set /p GO_CHECK=<"%TEMP%\tormentnexus_go_check.txt"
    if "!GO_CHECK!"=="200" (
        echo   Go sidecar already running on port 7778.
    ) else (
        echo   Starting Go sidecar on port 7778...
        start /B bin\tormentnexus.exe serve --port 7778 > nul 2>&1
    )
)

REM Start Next.js dashboard in background
set DASH_PORT=%TORMENTNEXUS_DASH_PORT%
if "%DASH_PORT%"=="" set DASH_PORT=3000
REM Fix Turbopack LevelDB deadlock on Windows
set NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%DASH_PORT%/dashboard > "%TEMP%\tormentnexus_web_check.txt" 2>nul
set /p WEB_CHECK=<"%TEMP%\tormentnexus_web_check.txt"
if "!WEB_CHECK!"=="200" (
    echo   Next.js dashboard already running on port %DASH_PORT%.
) else (
    echo   Starting Next.js dashboard on port %DASH_PORT%...
    start /B cmd /c "cd apps\web && set NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1 && npx next dev --port %DASH_PORT% > nul 2>&1"
)

REM Start TypeScript control plane
set TORMENTNEXUS_PORT=%TORMENTNEXUS_PORT%
if "%TORMENTNEXUS_PORT%"=="" set TORMENTNEXUS_PORT=4100

REM Wait for dashboard to be ready
echo   Waiting for dashboard to compile...
set WAIT_COUNT=0
:dash_wait_ts
if !WAIT_COUNT! GEQ 30 goto :dash_done_ts
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%DASH_PORT%/dashboard > "%TEMP%\tormentnexus_dash_wait.txt" 2>nul
set /p DASH_READY=<"%TEMP%\tormentnexus_dash_wait.txt"
if "!DASH_READY!"=="200" goto :dash_done_ts
    set /a WAIT_COUNT+=1
    timeout /t 2 > nul
    goto :dash_wait_ts
:dash_done_ts

REM Check if core is already running
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%TORMENTNEXUS_PORT%/health > "%TEMP%\tormentnexus_core_check.txt" 2>nul
set /p CORE_CHECK=<"%TEMP%\tormentnexus_core_check.txt"
if "!CORE_CHECK!"=="200" (
    echo   TS control plane already running on port %TORMENTNEXUS_PORT%.
) else (
    echo   Starting TS control plane on port %TORMENTNEXUS_PORT%...
)
echo.
echo   tRPC:      http://0.0.0.0:%TORMENTNEXUS_PORT%/trpc
echo   REST:      http://0.0.0.0:%TORMENTNEXUS_PORT%/api
echo   Go:        http://127.0.0.1:7778/health
echo   Dashboard: http://localhost:%DASH_PORT%/dashboard
echo.
if not "!CORE_CHECK!"=="200" (
    echo Press Ctrl+C to stop all services.
    echo.
    node packages/cli/dist/cli/src/index.js start --port %TORMENTNEXUS_PORT% !REMAINING_ARGS!
) else (
    echo Press Ctrl+C to stop all services.
    echo.
    echo All services running. Core was already active.
    timeout /t 99999 > nul
)
goto :eof

REM ===== GO-PRIMARY PATH =====
:go_primary

REM -- 1. Locate Go binary --------------------------
set GO_BIN=
if exist tormentnexus.exe set GO_BIN=tormentnexus.exe
if exist bin\tormentnexus.exe set GO_BIN=bin\tormentnexus.exe
if "!GO_BIN!"=="" (
    echo [FATAL] No Go binary found. Build with: cd go \&\& go build -o ..\tormentnexus.exe ./cmd/tormentnexus
    exit /b 1
)
echo [1/5] Using Go binary: !GO_BIN!

REM -- 2. Quick install (optional, for compat) ------ 
if /I "%TORMENTNEXUS_SKIP_INSTALL%"=="1" (
    echo [2/5] Skipping install (TORMENTNEXUS_SKIP_INSTALL=1)
) else (
    echo [2/5] Quick install (ignore-scripts)...
    call pnpm install --ignore-scripts 2>nul
)

REM -- 3. Start Go control plane on port 7778 ------
echo [3/5] Starting Go control plane...
set GO_PORT=7778
set GO_WAIT_COUNT=0

:go_health_check
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:!GO_PORT!/health > "%TEMP%\tormentnexus_go_check_!GO_PORT!.txt" 2>nul
set /p GO_HEALTH=<"%TEMP%\tormentnexus_go_check_!GO_PORT!.txt"
if "!GO_HEALTH!"=="200" (
    echo   Go control plane already running on port !GO_PORT!.
) else (
    if !GO_WAIT_COUNT! LSS 3 (
        start /B !GO_BIN! serve --port !GO_PORT! > nul 2>&1
        echo   Starting Go control plane on port !GO_PORT!...
        timeout /t 4 > nul
        set /a GO_WAIT_COUNT+=1
        curl -s -o nul -w "%%{http_code}" http://127.0.0.1:!GO_PORT!/health > "%TEMP%\tormentnexus_go_check_!GO_PORT!.txt" 2>nul
        set /p GO_HEALTH=<"%TEMP%\tormentnexus_go_check_!GO_PORT!.txt"
        if "!GO_HEALTH!"=="200" (
            echo   Go control plane started on port !GO_PORT!.
        ) else (
            if !GO_PORT!==7778 (
                echo   [WARN] Port 7778 unavailable, trying 7777...
                set GO_PORT=7777
                set GO_WAIT_COUNT=0
                goto :go_health_check
            ) else (
                echo   [WARN] Go control plane failed to start on ports 7778/7777.
                echo   Check if another process is using those ports.
            )
        )
    ) else (
        echo   [WARN] Go control plane not responding after 3 attempts.
    )
)

REM Export the actual Go port for subsequent steps
set TORMENTNEXUS_GO_PORT=!GO_PORT!

REM -- 4. Start Dashboard (compat mode) -------------
echo [4/5] Starting dashboard...
set DASH_PORT=%TORMENTNEXUS_DASH_PORT%
if "%DASH_PORT%"=="" set DASH_PORT=3000
set NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1
start /B cmd /c "cd apps\web && set NEXT_PRIVATE_DISABLE_TURBOPACK_CACHE=1 && npx next dev --port %DASH_PORT% > nul 2>&1"
echo   Dashboard starting on port %DASH_PORT% (Go-compat mode).

REM -- 5. Start TS compatibility (optional) ---------
echo [5/5] TS compatibility layer...
set TORMENTNEXUS_PORT=%TORMENTNEXUS_PORT%
if "%TORMENTNEXUS_PORT%"=="" set TORMENTNEXUS_PORT=4100

REM Check if TS is already running
curl -s -o nul -w "%%{http_code}" http://127.0.0.1:%TORMENTNEXUS_PORT%/health > "%TEMP%\tormentnexus_ts_check.txt" 2>nul
set /p TS_CHECK=<"%TEMP%\tormentnexus_ts_check.txt"
if "!TS_CHECK!"=="200" (
    echo   TS compatibility layer already running on port %TORMENTNEXUS_PORT%.
) else (
    echo   TS compatibility not running (optional, start with: tormentnexus start --runtime node)
)
echo.
echo   Go API:    http://127.0.0.1:7778/api
echo   Go Health: http://127.0.0.1:7778/health
echo   Dashboard: http://localhost:%DASH_PORT%/dashboard
echo   Runtime:   Go-primary
echo.
echo Press Ctrl+C to stop all services.
echo.
timeout /t 99999 > nul
