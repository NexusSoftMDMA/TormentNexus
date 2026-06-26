@echo off
setlocal
title BobbyBookmarks
cd /d "%~dp0"

echo [BobbyBookmarks] Starting...
python --version >nul 2>nul
if errorlevel 1 (
    echo [BobbyBookmarks] python not found. Please install it.
    pause
    exit /b 1
)

python -m bobbybookmarks

if errorlevel 1 (
    echo [BobbyBookmarks] Exited with error code %errorlevel%.
    pause
)
endlocal
