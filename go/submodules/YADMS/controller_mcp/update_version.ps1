param (
    [string]$VersionFile = "version.txt",
    [string]$CsFile = "VersionInfo.cs"
)

if (-not (Test-Path $VersionFile)) { Set-Content -Path $VersionFile -Value "1.0.0" }
$versionStr = (Get-Content $VersionFile).Trim()
$parts = $versionStr.Split('.')
if ($parts.Length -lt 3) { $parts = @(1, 0, 0) }

$major = [int]$parts[0]
$minor = [int]$parts[1]
$patch = [int]$parts[2]

$versionChanged = $false
if ($env:BumpMajor -eq "true") { 
    $major++; $minor = 0; $patch = 0
    $versionChanged = $true 
} elseif ($env:BumpMinor -eq "true") { 
    $minor++; $patch = 0
    $versionChanged = $true 
} elseif ($env:BumpPatch -eq "true") { 
    $patch++
    $versionChanged = $true 
}

$newVersion = "$major.$minor.$patch"

if ($versionChanged) {
    Set-Content -Path $VersionFile -Value $newVersion
    $csContent = "namespace controller_mcp { public static class VersionInfo { public const string CurrentVersion = `"$newVersion`"; } }"
    Set-Content -Path $CsFile -Value $csContent
    Write-Host "Auto-incremented version to $newVersion"
} else {
    # Ensure C# file exists even if no bump requested
    if (-not (Test-Path $CsFile)) {
        $csContent = "namespace controller_mcp { public static class VersionInfo { public const string CurrentVersion = `"$newVersion`"; } }"
        Set-Content -Path $CsFile -Value $csContent
    }
    Write-Host "Build complete for version $newVersion (No version bump requested)"
}

$assemblyInfo = "Properties\AssemblyInfo.cs"
if (Test-Path $assemblyInfo) {
    $content = Get-Content $assemblyInfo -Raw
    $content = $content -replace '\[assembly: AssemblyVersion\(".*"\)\]', "[assembly: AssemblyVersion(`"$newVersion.0`")]"
    $content = $content -replace '\[assembly: AssemblyFileVersion\(".*"\)\]', "[assembly: AssemblyFileVersion(`"$newVersion.0`")]"
    Set-Content -Path $assemblyInfo -Value $content -NoNewline
}
