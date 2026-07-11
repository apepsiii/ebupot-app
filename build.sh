#!/bin/bash
# ============================================================
#  build.sh — Cross-compile e-Bupot Portal untuk Linux
#
#  Cara pakai:
#    bash build.sh
#
#  Hasil: file binary "ebupot-app" (linux/amd64, self-contained)
# ============================================================

set -e

echo ""
echo "========================================"
echo "  e-Bupot Portal — Build"
echo "========================================"
echo ""

# Cross-compile untuk Linux amd64 (CGO disabled = pure-Go SQLite)
echo "[1/2] Cross-compile (linux/amd64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ebupot-app .

echo "  [OK] Binary: ebupot-app"
echo ""

# Tampilkan ukuran
SIZE=$(du -h ebupot-app | cut -f1)
echo "[2/2] Selesai!"
echo "  File    : ebupot-app"
echo "  Ukuran  : $SIZE"
echo "  Platform: linux/amd64"
echo "  Embed   : templates, public, config (self-contained)"
echo ""
echo "========================================"
echo "  Upload ke VPS:"
echo "    scp ebupot-app root@IP_VPS:/opt/ebupot-app/"
echo ""
echo "  Jalankan di VPS:"
echo "    cd /opt/ebupot-app && chmod +x ebupot-app && ./ebupot-app"
echo "    (wizard instalasi muncul otomatis)"
echo "========================================"
echo ""
