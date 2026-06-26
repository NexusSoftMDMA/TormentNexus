@echo off
REM Thin wrapper - all real logic lives in bin\cli.js.
REM Args are forwarded so `start.bat --port 4000 --api-port 9000` works.
cd /d "%~dp0"
node bin\cli.js %*
