# Install the latest Google Chrome enterprise build on Windows (amd64 or arm64).
$ErrorActionPreference = "Stop"

$isArm = ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') -or ($env:RUNNER_ARCH -eq 'ARM64')
if ($isArm) {
    $url = 'https://dl.google.com/tag/s/dl/chrome/install/googlechromestandaloneenterprise_arm64.msi'
} else {
    $url = 'https://dl.google.com/tag/s/dl/chrome/install/googlechromestandaloneenterprise64.msi'
}

$temp = if ($env:RUNNER_TEMP) { $env:RUNNER_TEMP } else { [System.IO.Path]::GetTempPath() }
$msi = Join-Path $temp 'google-chrome-latest.msi'
$userAgent = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36'

Write-Host "Downloading latest Chrome: $url"
Invoke-WebRequest -Uri $url -OutFile $msi -UseBasicParsing -UserAgent $userAgent

Write-Host "Installing Chrome silently..."
$proc = Start-Process msiexec.exe -ArgumentList @('/i', $msi, '/qn', '/norestart') -Wait -PassThru -NoNewWindow
if ($proc.ExitCode -ne 0) {
    throw "msiexec failed with exit code $($proc.ExitCode)"
}

$candidates = @(
    (Join-Path ${env:ProgramFiles} 'Google\Chrome\Application\chrome.exe'),
    (Join-Path ${env:ProgramFiles(x86)} 'Google\Chrome\Application\chrome.exe'),
    (Join-Path $env:LOCALAPPDATA 'Google\Chrome\Application\chrome.exe')
)

$chrome = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (-not $chrome) {
    throw "Chrome not found after install. Checked:`n  $($candidates -join "`n  ")"
}

$env:CHROME_EXE = $chrome
"CHROME_EXE=$chrome" >> $env:GITHUB_ENV

$version = (Get-Item $chrome).VersionInfo.ProductVersion
Write-Host "Chrome $version installed at $chrome"
