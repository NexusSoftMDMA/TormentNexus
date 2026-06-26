@echo off
echo Starting Enterprise Sales Bot...
if not exist "bin/sales_bot" (
    echo Binary not found. Running build...
    call build.bat
)
bin\sales_bot %*
