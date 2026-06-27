# Build Chrome.exe (Windows amd64 / arm64)
param(
    [ValidateSet("amd64", "arm64", "all")]
    [string[]]$Arch = @("all")
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $root

go run ./tools/genicon

$targets = if ($Arch -contains "all") { @("amd64", "arm64") } else { $Arch }
New-Item -ItemType Directory -Force -Path "dist" | Out-Null

foreach ($goarch in $targets) {
    $out = Join-Path $root ("dist\Chrome-{0}.exe" -f $goarch)
    Write-Host "build $goarch -> $out"
    $env:GOOS = "windows"
    $env:GOARCH = $goarch
    go build -ldflags "-H=windowsgui" -o $out ./cmd/chrome
}

Write-Host "done"
