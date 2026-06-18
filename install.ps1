# install.ps1
$ErrorActionPreference = "Stop"

$Repo = if ($env:WENET_CLI_REPO) { $env:WENET_CLI_REPO } else { "wenet-ec/wenet-cli" }
$Version = if ($env:WENET_CLI_VERSION) { $env:WENET_CLI_VERSION } else { "latest" }
$InstallDir = if ($env:WENET_CLI_INSTALL_DIR) { $env:WENET_CLI_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\wenet" }

$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
}

if ($Version -eq "latest") {
    $Url = "https://github.com/$Repo/releases/latest/download/wenet_windows_$Arch.zip"
} else {
    $Url = "https://github.com/$Repo/releases/download/$Version/wenet_windows_$Arch.zip"
}

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TempDir | Out-Null

try {
    $Archive = Join-Path $TempDir "wenet.zip"
    Invoke-WebRequest -Uri $Url -OutFile $Archive
    Expand-Archive -Path $Archive -DestinationPath $TempDir -Force

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path (Join-Path $TempDir "wenet.exe") -Destination (Join-Path $InstallDir "wenet.exe") -Force

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (($UserPath -split ";") -notcontains $InstallDir) {
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        Write-Host "Added $InstallDir to your user PATH. Open a new terminal before running wenet."
    }

    Write-Host "wenet installed to $(Join-Path $InstallDir "wenet.exe")"
} finally {
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
}
