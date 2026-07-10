# ============================================================
#  deploy.ps1 — Deploy e-Bupot Portal ke VPS (Binary + Systemd)
#
#  Cara pakai:
#    .\deploy\deploy.ps1 -VpsHost user@IP_VPS -Domain ebupot.domainanda.com
#
#  Opsi:
#    -VpsHost       SSH target (wajib), contoh: root@192.168.1.10
#    -Domain        Domain publik (wajib), contoh: ebupot.domainanda.com
#    -RemotePath    Path instalasi di VPS (default: /opt/ebupot-app)
#    -SshKey        Path SSH private key (opsional, contoh: ~/.ssh/id_rsa)
#    -SkipSSL       Skip setup SSL/Let's Encrypt (untuk testing)
# ============================================================

param(
    [Parameter(Mandatory=$true)]
    [string]$VpsHost,
    [Parameter(Mandatory=$true)]
    [string]$Domain,
    [string]$RemotePath = "/opt/ebupot-app",
    [string]$SshKey = "",
    [switch]$SkipSSL
)

$ErrorActionPreference = "Stop"

# --- SSH/SCP args ---
$sshArgs = @()
$scpArgs = @()
if ($SshKey -ne "") {
    $sshArgs += @("-i", $SshKey)
    $scpArgs += @("-i", $SshKey)
}

function Invoke-SSH($command) {
    Write-Host "  [SSH] $command" -ForegroundColor DarkGray
    ssh @sshArgs $VpsHost $command
}
function Invoke-SCP($local, $remote) {
    Write-Host "  [SCP] $local -> $remote" -ForegroundColor DarkGray
    scp @scpArgs -r $local "${VpsHost}:${remote}"
}

# ============================================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  e-Bupot Portal — Deployment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  VPS Host : $VpsHost"
Write-Host "  Domain   : $Domain"
Write-Host "  Path     : $RemotePath"
Write-Host "  SSL      : $(if ($SkipSSL) { 'Skip' } else { 'Yes (Let'"'"'s Encrypt)' })"
Write-Host ""

# ============================================================
# 1. Cross-compile binary untuk Linux amd64
# ============================================================
Write-Host "[1/5] Cross-compile binary (linux/amd64)..." -ForegroundColor Yellow

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

go build -ldflags="-s -w" -o ebupot-linux .
if ($LASTEXITCODE -ne 0) {
    Write-Host "  BUILD GAGAL!" -ForegroundColor Red
    exit 1
}
$binSize = [math]::Round((Get-Item ebupot-linux).Length / 1MB, 1)
Write-Host "  Binary: ebupot-linux ($binSize MB)" -ForegroundColor Green

# Reset env
$env:GOOS = ""
$env:GOARCH = ""
$env:CGO_ENABLED = ""

# ============================================================
# 2. Upload file ke VPS
# ============================================================
Write-Host "`n[2/5] Upload file ke VPS..." -ForegroundColor Yellow

# Buat folder di VPS
Invoke-SSH "sudo mkdir -p $RemotePath && sudo chown -R `$USER:`$USER $RemotePath"

# Upload binary
Invoke-SCP "ebupot-linux" "$RemotePath/ebupot-app"

# Upload folder templates & public
Invoke-SCP "templates" "$RemotePath/"
Invoke-SCP "public" "$RemotePath/"

# Upload config files
Invoke-SCP "config.yaml" "$RemotePath/"
Invoke-SCP ".env.example" "$RemotePath/"
Invoke-SCP "deploy/install.sh" "$RemotePath/install.sh"

Write-Host "  Upload selesai" -ForegroundColor Green

# ============================================================
# 3. Jalankan install.sh di VPS
# ============================================================
Write-Host "`n[3/5] Setup aplikasi & systemd di VPS..." -ForegroundColor Yellow

$skipSslFlag = ""
if ($SkipSSL) { $skipSslFlag = "--skip-ssl" }

Invoke-SSH "chmod +x $RemotePath/install.sh && bash $RemotePath/install.sh --domain $Domain --path $RemotePath $skipSslFlag"

# ============================================================
# 4. Cleanup
# ============================================================
Write-Host "`n[4/5] Cleanup..." -ForegroundColor Yellow
Remove-Item ebupot-linux -Force -ErrorAction SilentlyContinue
Write-Host "  Binary lokal dihapus" -ForegroundColor Green

# ============================================================
# 5. Selesai
# ============================================================
Write-Host "`n[5/5] Deployment selesai!" -ForegroundColor Green
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Akses: https://$Domain" -ForegroundColor Cyan
Write-Host "  Login: admin / admin123" -ForegroundColor Cyan
Write-Host "  Log  : ssh $VpsHost 'sudo journalctl -u ebupot -f'" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "JANGAN LUPA:" -ForegroundColor Red
Write-Host "  1. Ganti password admin setelah login pertama" -ForegroundColor White
Write-Host "  2. Edit SESSION_SECRET di $RemotePath/.env" -ForegroundColor White
Write-Host "  3. Restart: ssh $VpsHost 'sudo systemctl restart ebupot'" -ForegroundColor White
Write-Host ""
