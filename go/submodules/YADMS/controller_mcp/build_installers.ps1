$ErrorActionPreference = "Stop"

Write-Host "========================================"
Write-Host " Generating Inno Setup Installers (4 SKUs)"
Write-Host "========================================"

$skus = @(
    "Lite_Standard",
    "Lite_Hacked",
    "Bundled_Standard",
    "Bundled_Hacked"
)

$isccPath = "iscc"
if (-not (Get-Command $isccPath -ErrorAction SilentlyContinue)) {
    # Fallback to hardcoded standard directory if not in PATH
    $isccPath = "C:\Program Files (x86)\Inno Setup 6\iscc.exe"
    if (-not (Test-Path $isccPath)) {
        Write-Error "Inno Setup Compiler (iscc.exe) not found in PATH or at $isccPath"
        exit 1
    }
}

$version = (Get-Content "version.txt" -Raw).Trim()

# Copy license to txt for Inno Setup
Copy-Item "..\LICENSE" -Destination "LICENSE.txt" -Force
Copy-Item "..\SOCIALS.txt" -Destination "SOCIALS.txt" -Force

foreach ($sku in $skus) {
    Write-Host "Compiling Installer for $sku (Version $version)..."
    & $isccPath "/DAppVersion=$version" "/DSkuName=$sku" "installer.iss"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build installer for $sku"
        exit 1
    }
}

Write-Host "========================================"
Write-Host " All Installers Generated Successfully!"
Write-Host "========================================"
