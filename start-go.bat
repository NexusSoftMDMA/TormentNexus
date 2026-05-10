@echo off
setlocal EnableDelayedExpansion

REM ==============================================
REM  HyperCode AIOS - Go Sidecar Startup
REM  Starts the Go-native control plane only
REM ==============================================

for /f "tokens=*" %%v in ('type VERSION 2^>nul') do set VER=%%v
if "%VER%"=="" set VER=dev

REM -- Configuration ----------------------------------
set GO_PORT=%HYPERCODE_GO_PORT%
if "%GO_PORT%"=="" set GO_PORT=4300
set GO_HOST=%HYPERCODE_GO_HOST%
if "%GO_HOST%"=="" set GO_HOST=127.0.0.1
set GO_CONFIG=%HYPERCODE_GO_CONFIG%
if "%GO_CONFIG%"=="" set GO_CONFIG=%USERPROFILE%\.hypercode-go

REM -- 1. Check Go Installation ----------------------
where go >nul 2>nul
if errorlevel 1 (
    echo [FATAL] Go not found. Install Go 1.22+ from https://go.dev/dl/
    exit /b 1
)

REM -- 2. Build Go Sidecar ---------------------------
echo [1/3] Building Go sidecar v%VER%...
cd go
go build -ldflags "-X github.com/hypercodehq/hypercode-go/internal/buildinfo.Version=%VER%" -buildvcs=false -o ..\bin\hypercode.exe ./cmd/hypercode 2>nul
if errorlevel 1 (
    echo [FATAL] Go build failed.
    cd ..
    exit /b 1
)
echo   OK bin\hypercode.exe
cd ..

REM -- 3. Check if Already Running -------------------
curl -s -o nul -w "%%{http_code}" http://%GO_HOST%:%GO_PORT%/health > "%TEMP%\hc_go_check.txt" 2>nul
set /p GO_CHECK=<"%TEMP%\hc_go_check.txt"
if "!GO_CHECK!"=="200" (
    echo [2/3] Go sidecar already running on %GO_HOST%:%GO_PORT%.
    echo.
    echo   API:      http://%GO_HOST%:%GO_PORT%/api/index
    echo   Health:   http://%GO_HOST%:%GO_PORT%/health
    echo   Version:  http://%GO_HOST%:%GO_PORT%/version
    echo.
    echo Press Ctrl+C to exit.
    timeout /t 99999 > nul
    exit /b 0
)

REM -- 4. Launch Go Sidecar --------------------------
echo [2/3] Starting Go sidecar on %GO_HOST%:%GO_PORT%...
echo.
echo   API:      http://%GO_HOST%:%GO_PORT%/api/index
echo   Health:   http://%GO_HOST%:%GO_PORT%/health
echo   Version:  http://%GO_HOST%:%GO_PORT%/version
echo   Config:   %GO_CONFIG%
echo.
echo [3/3] Running (Ctrl+C to stop)...
echo.
bin\hypercode.exe -port %GO_PORT% -host %GO_HOST% -config-dir %GO_CONFIG%
exit /b %errorlevel%
