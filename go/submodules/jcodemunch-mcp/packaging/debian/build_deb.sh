#!/bin/bash
# Build a Debian/Ubuntu .deb package for jcodemunch-mcp.
# Contributed by Tikilou — tested on Debian 13 / Proxmox LXC containers.
#
# Usage:
#   cd /path/to/jcodemunch-mcp
#   bash packaging/debian/build_deb.sh
#
# Output: jcodemunch-mcp_<version>_all.deb in the current directory.
# Install: sudo apt install ./jcodemunch-mcp_<version>_all.deb
# Enable:  sudo systemctl enable --now jcodemunch-mcp

set -e

# --- Package Configuration ---
PACKAGE_NAME="jcodemunch-mcp"
SERVICE_USER="jcodemunch"
MAINTAINER="Tikilou <tikilou@local>"
DESCRIPTION="Optimized MCP server for code exploration (Debian standard)"
BUILD_DIR="debian_build"
INSTALL_PATH="/opt/$PACKAGE_NAME"
DATA_PATH="/var/lib/$PACKAGE_NAME"
CONFIG_PATH="/etc/$PACKAGE_NAME"

# 1. Check build prerequisites
for cmd in dpkg-deb python3 pip; do
    if ! command -v $cmd &> /dev/null; then
        echo "Error: Build tools (dpkg-dev, python3-venv) are required."
        exit 1
    fi
done

# Separate venv check because it is a python3 module
if ! python3 -m venv --help &> /dev/null; then
    echo "Error: python3-venv is required."
    exit 1
fi

# 2. Extract version
VERSION=$(grep -m 1 "version =" pyproject.toml | cut -d '"' -f 2)
if [ -z "$VERSION" ]; then VERSION="1.0.0"; fi
echo "Building package $PACKAGE_NAME v$VERSION for Debian 13..."

# 3. Prepare directory tree
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR/DEBIAN"
mkdir -p "$BUILD_DIR$INSTALL_PATH"
mkdir -p "$BUILD_DIR/usr/bin"
mkdir -p "$BUILD_DIR$CONFIG_PATH"
mkdir -p "$BUILD_DIR$DATA_PATH"
mkdir -p "$BUILD_DIR/lib/systemd/system"

# 4. Create VirtualEnv (dependency isolation)
echo "Creating virtual environment in $INSTALL_PATH..."
python3 -m venv "$BUILD_DIR$INSTALL_PATH"
"$BUILD_DIR$INSTALL_PATH/bin/pip" install --upgrade pip
"$BUILD_DIR$INSTALL_PATH/bin/pip" install ".[http,watch,anthropic,gemini,openai,dbt]"

# Replace absolute build paths with the final install path in venv scripts
BUILD_ABS_PATH=$(realpath "$BUILD_DIR$INSTALL_PATH")
find "$BUILD_DIR$INSTALL_PATH/bin" -type f -exec sed -i "s|$BUILD_ABS_PATH|/opt/$PACKAGE_NAME|g" {} +

# 5. Default configuration (Standard SSE for network access)
cat <<EOF > "$BUILD_DIR$CONFIG_PATH/config.jsonc"
{
  "transport": "streamable-http",
  "host": "0.0.0.0",
  "port": 8901,
  "log_level": "INFO",
  "share_savings": true,
  "discovery_hint": true
}
EOF

# 6. Executable wrapper
cat <<EOF > "$BUILD_DIR/usr/bin/$PACKAGE_NAME"
#!/bin/bash
# Define the storage path for data and config
export CODE_INDEX_PATH="$DATA_PATH"
# Load config from /etc if it is not already present in /var/lib
if [ ! -f "\$CODE_INDEX_PATH/config.jsonc" ] && [ ! -L "\$CODE_INDEX_PATH/config.jsonc" ]; then
    ln -sf "$CONFIG_PATH/config.jsonc" "\$CODE_INDEX_PATH/config.jsonc"
fi
exec "$INSTALL_PATH/bin/jcodemunch-mcp" "\$@"
EOF
chmod +x "$BUILD_DIR/usr/bin/$PACKAGE_NAME"

# 7. Systemd unit
cat <<EOF > "$BUILD_DIR/lib/systemd/system/$PACKAGE_NAME.service"
[Unit]
Description=jCodeMunch MCP Server
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
Environment=CODE_INDEX_PATH=$DATA_PATH
# Start in Streamable HTTP mode (modern MCP protocol, Antigravity-compatible)
ExecStart=/usr/bin/$PACKAGE_NAME serve --transport streamable-http --host 0.0.0.0 --port 8901
Restart=always
RestartSec=5
# Security: filesystem restrictions
ReadWritePaths=$DATA_PATH
CapabilityBoundingSet=
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
EOF

# 8. Debian scripts (Post-inst / Prerm)
cat <<EOF > "$BUILD_DIR/DEBIAN/postinst"
#!/bin/bash
set -e
# Create system user if it does not already exist
if ! id "$SERVICE_USER" >/dev/null 2>&1; then
    useradd --system --home-dir $DATA_PATH --shell /usr/sbin/nologin $SERVICE_USER
fi

# Manage permissions on the data directory
mkdir -p $DATA_PATH
chown -R $SERVICE_USER:$SERVICE_USER $DATA_PATH
chmod 750 $DATA_PATH

# Symlink config to /etc if not already done
if [ ! -L "$DATA_PATH/config.jsonc" ] && [ ! -f "$DATA_PATH/config.jsonc" ]; then
    ln -sf $CONFIG_PATH/config.jsonc $DATA_PATH/config.jsonc
    chown -h $SERVICE_USER:$SERVICE_USER $DATA_PATH/config.jsonc
fi

# Reload systemd
if [ -d /run/systemd/system ]; then
    systemctl daemon-reload
fi

echo "Installation completed."
echo "Enable the service with: sudo systemctl enable --now $PACKAGE_NAME"
EOF

cat <<EOF > "$BUILD_DIR/DEBIAN/prerm"
#!/bin/bash
set -e
if [ "\$1" = "remove" ] || [ "\$1" = "upgrade" ]; then
    if systemctl is-active --quiet $PACKAGE_NAME; then
        systemctl stop $PACKAGE_NAME || true
    fi
fi
EOF
chmod 755 "$BUILD_DIR/DEBIAN/postinst" "$BUILD_DIR/DEBIAN/prerm"

# 9. Control file
cat <<EOF > "$BUILD_DIR/DEBIAN/control"
Package: $PACKAGE_NAME
Version: $VERSION
Section: utils
Priority: optional
Architecture: all
Maintainer: $MAINTAINER
Depends: python3, python3-venv
Description: $DESCRIPTION
 An MCP server for ultra-fast code exploration.
 Installed in a standardized way for Debian 13 (Bookworm+).
 Isolated in $INSTALL_PATH.
EOF

# 10. Conffiles file (to avoid overwriting user changes during upgrades)
cat <<EOF > "$BUILD_DIR/DEBIAN/conffiles"
$CONFIG_PATH/config.jsonc
EOF

# 11. Final build
dpkg-deb --build "$BUILD_DIR" "${PACKAGE_NAME}_${VERSION}_all.deb"
rm -rf "$BUILD_DIR"

echo "------------------------------------------------------"
echo "SUCCESS: ${PACKAGE_NAME}_${VERSION}_all.deb generated."
echo "To install it: sudo apt install ./${PACKAGE_NAME}_${VERSION}_all.deb"
echo "------------------------------------------------------"
