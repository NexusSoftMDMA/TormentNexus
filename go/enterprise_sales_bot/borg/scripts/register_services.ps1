# TormentNexus Service Registration
# Run this script to register auto-start scheduled tasks

$action1 = New-ScheduledTaskAction -Execute "C:\Users\hyper\workspace\tormentnexus\tormentnexus.exe" -Argument "serve" -WorkingDirectory "C:\Users\hyper\workspace\tormentnexus"
$action2 = New-ScheduledTaskAction -Execute "C:\Program Files\nodejs\node.exe" -Argument "C:\Users\hyper\workspace\tormentnexus\apps\web\node_modules\.bin\next.cmd dev -p 7779" -WorkingDirectory "C:\Users\hyper\workspace\tormentnexus\apps\web"
$action3 = New-ScheduledTaskAction -Execute "pythonw" -Argument "-u C:\Users\hyper\workspace\tormentnexus\watchdog.py" -WorkingDirectory "C:\Users\hyper\workspace\tormentnexus"

$trigger = New-ScheduledTaskTrigger -AtLogOn
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable

try {
    Register-ScheduledTask -TaskName "TormentNexus Sidecar" -Action $action1 -Trigger $trigger -Settings $settings -Force
    Write-Host "✅ TormentNexus Sidecar scheduled task created"
} catch {
    Write-Host "❌ Sidecar: $_"
}

try {
    Register-ScheduledTask -TaskName "TormentNexus Dashboard" -Action $action2 -Trigger $trigger -Settings $settings -Force
    Write-Host "✅ TormentNexus Dashboard scheduled task created"
} catch {
    Write-Host "❌ Dashboard: $_"
}

try {
    Register-ScheduledTask -TaskName "TormentNexus Watchdog" -Action $action3 -Trigger $trigger -Settings $settings -Force
    Write-Host "✅ TormentNexus Watchdog scheduled task created"
} catch {
    Write-Host "❌ Watchdog: $_"
}

Write-Host ""
Write-Host "Done. Tasks will start on next login."
Write-Host "To run now: Start-ScheduledTask -TaskName 'TormentNexus Sidecar'"
