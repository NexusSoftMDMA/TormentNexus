@echo off
cd /d "C:\Users\hyper\workspace\tormentnexus"
echo ========================================
echo  TormentNexus Service Registration
echo  Run this as Administrator!
echo ========================================
echo.

echo Registering Go Sidecar (port 7778)...
sc create "TormentNexusSidecar" binPath="\"C:\Users\hyper\workspace\tormentnexus\tormentnexus.exe\" serve" start=auto displayname="TormentNexus Sidecar"
if %errorlevel%==0 (echo ✅) else (echo ⚠️ may already exist)
echo.

echo Registering Dashboard (port 7779)...
sc create "TormentNexusDashboard" binPath="\"C:\Program Files\nodejs\node.exe\" \"C:\Users\hyper\workspace\tormentnexus\apps\web\node_modules\.bin\next.cmd\" dev -p 7779" start=auto displayname="TormentNexus Dashboard"
if %errorlevel%==0 (echo ✅) else (echo ⚠️ may already exist)
echo.

echo Registering Watchdog...
sc create "TormentNexusWatchdog" binPath="\"C:\Python314\pythonw.exe\" -u \"C:\Users\hyper\workspace\tormentnexus\watchdog.py\"" start=auto displayname="TormentNexus Watchdog"
if %errorlevel%==0 (echo ✅) else (echo ⚠️ may already exist)
echo.

echo ========================================
echo  Done! Starting services...
echo ========================================
sc start TormentNexusSidecar
sc start TormentNexusDashboard
sc start TormentNexusWatchdog

echo.
echo Services registered and starting.
pause
