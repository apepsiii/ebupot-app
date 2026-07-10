#!/bin/bash
# ============================================================
#  install.sh — Setup e-Bupot Portal di VPS (systemd + Nginx + SSL)
#
#  Cara pakai:
#    bash install.sh --domain ebupot.domainanda.com --path /opt/ebupot-app
#
#  Opsi:
#    --domain   Domain publik (wajib)
#    --path     Path instalasi (default: /opt/ebupot-app)
#    --skip-ssl Skip setup SSL/Let's Encrypt
# ============================================================

set -e

DOMAIN=""
APP_PATH="/opt/ebupot-app"
SKIP_SSL=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --domain)  DOMAIN="$2"; shift 2 ;;
        --path)    APP_PATH="$2"; shift 2 ;;
        --skip-ssl) SKIP_SSL=true; shift ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

if [ -z "$DOMAIN" ]; then
    echo "ERROR: --domain wajib diisi"
    echo "Contoh: bash install.sh --domain ebupot.domainanda.com"
    exit 1
fi

echo ""
echo "========================================"
echo "  e-Bupot Portal — VPS Setup"
echo "========================================"
echo "  Domain : $DOMAIN"
echo "  Path   : $APP_PATH"
echo "  SSL    : $([ "$SKIP_SSL" = true ] && echo 'Skip' || echo 'Yes')"
echo ""

# ============================================================
# 1. Buat folder data & uploads
# ============================================================
echo "[1/6] Membuat folder data & uploads..."
mkdir -p "$APP_PATH/data"
mkdir -p "$APP_PATH/uploads/ebupots"
touch "$APP_PATH/data/.gitkeep" "$APP_PATH/uploads/.gitkeep"
chmod -R 755 "$APP_PATH"
echo "  OK"

# ============================================================
# 2. Setup .env production
# ============================================================
echo "[2/6] Setup .env production..."

# Generate random session secret
SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | xxd -p 2>/dev/null || echo "change-this-secret")

if [ ! -f "$APP_PATH/.env" ]; then
    cat > "$APP_PATH/.env" <<EOF
# e-Bupot Portal — Production
APP_ENV=production
SERVER_HOST=127.0.0.1
SERVER_PORT=8080
SERVER_DOMAIN=$DOMAIN
DB_PATH=data/ebupot.db
UPLOAD_MAX_SIZE_MB=100
QR_RECOVERY_LEVEL=high
QR_SIZE=512
SESSION_SECRET=$SECRET
SESSION_MAX_AGE=86400
EOF
    echo "  .env dibuat dengan SESSION_SECRET random"
else
    echo "  .env sudah ada, skip (edit manual jika perlu)"
fi

# ============================================================
# 3. Buat systemd service
# ============================================================
echo "[3/6] Setup systemd service..."

# Deteksi user yang menjalankan script
CURRENT_USER=$(whoami)
if [ "$CURRENT_USER" = "root" ]; then
    SERVICE_USER="www-data"
else
    SERVICE_USER="$CURRENT_USER"
fi

# Pastikan user www-data ada jika dipakai
if [ "$SERVICE_USER" = "www-data" ]; then
    id -u www-data &>/dev/null || useradd -r -s /usr/sbin/nologin www-data
    chown -R www-data:www-data "$APP_PATH/data" "$APP_PATH/uploads"
fi

cat > /tmp/ebupot.service <<EOF
[Unit]
Description=e-Bupot Portal & QR Generator
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$APP_PATH
ExecStart=$APP_PATH/ebupot-app
EnvironmentFile=$APP_PATH/.env
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=ebupot
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$APP_PATH/data $APP_PATH/uploads

[Install]
WantedBy=multi-user.target
EOF

sudo cp /tmp/ebupot.service /etc/systemd/system/ebupot.service
sudo systemctl daemon-reload
sudo systemctl enable ebupot
sudo systemctl restart ebupot
rm /tmp/ebupot.service

sleep 2
if systemctl is-active --quiet ebupot; then
    echo "  Service berjalan (active)"
else
    echo "  WARNING: Service belum active, cek: sudo journalctl -u ebupot -e"
fi

# ============================================================
# 4. Install & setup Nginx
# ============================================================
echo "[4/6] Setup Nginx reverse proxy..."

if ! command -v nginx &>/dev/null; then
    echo "  Installing Nginx..."
    if command -v apt &>/dev/null; then
        sudo apt update -qq && sudo apt install -y -qq nginx
    elif command -v yum &>/dev/null; then
        sudo yum install -y nginx
    else
        echo "  ERROR: Tidak bisa install Nginx otomatis. Install manual."
    fi
fi

cat > /tmp/ebupot-nginx <<EOF
server {
    listen 80;
    server_name $DOMAIN;
    client_max_body_size 100M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }
}
EOF

sudo cp /tmp/ebupot-nginx /etc/nginx/sites-available/ebupot 2>/dev/null || sudo cp /tmp/ebupot-nginx /etc/nginx/conf.d/ebupot.conf
if [ -d /etc/nginx/sites-enabled ]; then
    sudo ln -sf /etc/nginx/sites-available/ebupot /etc/nginx/sites-enabled/ebupot
fi
rm /tmp/ebupot-nginx

# Hapus default config jika konflik
sudo rm -f /etc/nginx/sites-enabled/default 2>/dev/null

sudo nginx -t 2>/dev/null && sudo systemctl reload nginx || echo "  WARNING: Nginx config test gagal, cek manual"
echo "  Nginx configured untuk $DOMAIN"

# ============================================================
# 5. Setup SSL (Let's Encrypt)
# ============================================================
if [ "$SKIP_SSL" = false ]; then
    echo "[5/6] Setup SSL (Let's Encrypt)..."

    if ! command -v certbot &>/dev/null; then
        echo "  Installing Certbot..."
        if command -v apt &>/dev/null; then
            sudo apt install -y -qq certbot python3-certbot-nginx
        else
            echo "  Skip: Certbot tidak bisa diinstall otomatis"
        fi
    fi

    if command -v certbot &>/dev/null; then
        sudo certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --register-unsafely-without-email --redirect || \
        echo "  WARNING: Certbot gagal. Pastikan domain $DOMAIN sudah mengarah ke IP VPS ini."
        echo "  Jalankan manual: sudo certbot --nginx -d $DOMAIN"
    else
        echo "  Skip: Install certbot manual: sudo apt install certbot python3-certbot-nginx"
    fi
else
    echo "[5/6] SSL di-skip (--skip-ssl)"
fi

# ============================================================
# 6. Selesai
# ============================================================
echo "[6/6] Selesai!"
echo ""
echo "========================================"
echo "  Deployment Selesai!"
echo "========================================"
echo ""
echo "  URL    : http://$DOMAIN"
if [ "$SKIP_SSL" = false ]; then
echo "  URL SSL: https://$DOMAIN"
fi
echo "  Login  : admin / admin123"
echo ""
echo "  Log aplikasi : sudo journalctl -u ebupot -f"
echo "  Restart      : sudo systemctl restart ebupot"
echo "  Status       : sudo systemctl status ebupot"
echo "  Edit config  : nano $APP_PATH/.env"
echo ""
echo "  JANGAN LUPA:"
echo "    1. Ganti password admin setelah login pertama"
echo "    2. Pastikan DNS domain $DOMAIN mengarah ke IP VPS ini"
echo ""
