# ============================================================
#  deploy.ps1 — Deploy e-Bupot Portal ke VPS (Single Binary)
#
#  Binary sudah embed semua (templates, public, config).
#  Cukup upload 1 file, jalankan, wizard otomatis muncul.
#
#  Cara pakai:
#    .\deploy\deploy.ps1 -VpsHost root@IP_VPS
#
#  Opsi:
#    -VpsHost    SSH target (wajib), contoh: root@192.168.1.10
#    -RemotePath Path di VPS (default: /opt/ebupot-app)
#    -SshKey     Path SSH private key (opsional)
# ============================================================

param(
    [Parameter(Mandatory=$true)]
    [string]$VpsHost,
    [string]$RemotePath = "/opt/ebupot-app",
    [string]$SshKey = ""
)

$ErrorActionPreference = "Stop"

$sshArgs = @()
if ($SshKey -ne "") { $sshArgs += @("-i", $SshKey) }

function Invoke-SSH($command) {
    Write-Host "  [SSH] $command" -ForegroundColor DarkGray
    ssh @sshArgs $VpsHost $command
}
function Invoke-SCP($local, $remote) {
    Write-Host "  [SCP] $local -> $remote" -ForegroundColor DarkGray
    scp @sshArgs $local "${VpsHost}:${remote}"
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  e-Bupot Portal — Deploy (Single Binary)" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  VPS Host : $VpsHost" -ForegroundColor White
Write-Host "  Path     : $RemotePath" -ForegroundColor White
Write-Host ""

# ============================================================
# 1. Cross-compile binary untuk Linux amd64
# ============================================================
Write-Host "[1/3] Cross-compile binary (linux/amd64)..." -ForegroundColor Yellow

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

go build -ldflags="-s -w" -o ebupot-linux .
if ($LASTEXITCODE -ne 0) {
    Write-Host "  BUILD GAGAL!" -ForegroundColor Red
    exit 1
}
$binSize = [math]::Round((Get-Item ebupot-linux).Length / 1MB, 1)
Write-Host "  Binary: ebupot-linux ($binSize MB) — semua file ter-embed" -ForegroundColor Green

$env:GOOS = ""
$env:GOARCH = ""
$env:CGO_ENABLED = ""

# ============================================================
# 2. Upload single binary ke VPS
# ============================================================
Write-Host "`n[2/3] Upload binary ke VPS..." -ForegroundColor Yellow

Invoke-SSH "sudo mkdir -p $RemotePath && sudo chown -R ``$USER:``$USER $RemotePath"
Invoke-SCP "ebupot-linux" "$RemotePath/ebupot-app"
Invoke-SSH "chmod +x $RemotePath/ebupot-app"

Write-Host "  Upload selesai" -ForegroundColor Green

# ============================================================
# 3. Jalankan binary di VPS (wizard otomatis muncul)
# ============================================================
Write-Host "`n[3/3] Jalankan aplikasi di VPS..." -ForegroundColor Yellow
Write-Host "  Binary akan membuka setup wizard interaktif." -ForegroundColor White
Write-Host "  Jawab pertanyaan (port, domain, dll) atau tekan Enter untuk default." -ForegroundColor White
Write-Host ""
Write-Host "  Jalankan manual di VPS:" -ForegroundColor Cyan
Write-Host "    ssh $VpsHost" -ForegroundColor White
Write-Host "    cd $RemotePath && ./ebupot-app" -ForegroundColor White
Write-Host ""

# Cleanup
Remove-Item ebupot-linux -Force -ErrorAction SilentlyContinue

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Binary ter-upload ke $RemotePath/ebupot-app" -ForegroundColor Cyan
Write-Host "" -ForegroundColor White
Write-Host "  Langkah selanjutnya (di VPS):" -ForegroundColor Yellow
Write-Host "    cd $RemotePath" -ForegroundColor White
Write-Host "    ./ebupot-app    # wizard akan muncul otomatis" -ForegroundColor White
Write-Host "" -ForegroundColor White
Write-Host "  Setelah wizard selesai, server langsung jalan." -ForegroundColor Green
Write-Host "  Untuk start ulang: ./ebupot-app" -ForegroundColor White
Write-Host "========================================" -ForegroundColor Cyan
