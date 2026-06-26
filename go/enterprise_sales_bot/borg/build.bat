@echo off
setlocal

echo Building TormentNexus TORMENTNEXUS...

REM Build Go sidecar
where go >nul 2>nul
if errorlevel 1 (
    echo [SKIP] Go not found
    goto :ts_build
)

echo [1/2] Building Go sidecar...
for /f "tokens=*" %%v in ('type VERSION') do set VER=%%v
cd go
go build -ldflags "-s -w -X github.com/tormentnexushq/tormentnexus-go/internal/buildinfo.Version=%VER%" -buildvcs=false -o ..\bin\tormentnexus.exe ./cmd/tormentnexus
if errorlevel 1 (
    echo [FAIL] Go build failed
    cd ..
    exit /b 1
)
cd ..
echo OK - bin\tormentnexus.exe built

:ts_build
REM Build TypeScript core
where npx >nul 2>nul
if errorlevel 1 (
    echo [SKIP] npx not found
    goto :done
)

echo [2/2] Building TypeScript core...
cd packages\core
call npx tsc --project tsconfig.json
if errorlevel 1 (
    echo [FAIL] TypeScript build failed
    cd ..\..
    exit /b 1
)
cd ..\..
echo OK - packages\core built

:done
echo.
echo Build complete.
