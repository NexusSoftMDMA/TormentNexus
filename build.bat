@echo off
setlocal

echo Building TormentNexus TORMENTNEXUS...

REM Build Go sidecar
where go >nul 2>nul
if errorlevel 1 (
    echo [SKIP] Go not found
    goto :done
)

echo Building Go sidecar...
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

:done
echo.
echo Build complete.
