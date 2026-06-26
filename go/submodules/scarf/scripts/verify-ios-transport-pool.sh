#!/usr/bin/env bash
#
# Verify the iOS transport connection pooling + open-coalescing (gh#112)
# against a REAL sshd — the churn that broke ScarfGo chat-init only reproduces
# with an actual SSH server, which unit tests can't provide.
#
# Self-contained: spins up an ephemeral, localhost-only sshd on a high port
# with a throwaway host key / config / authorized_keys (NO system Remote Login,
# no sudo), runs the gated live integration test, then tears everything down.
# All writes go to a throwaway HERMES_HOME — never the real ~/.hermes.
#
# Requires `hermes` installed locally (the test invokes it over the SSH loop).
#
# Usage: ./scripts/verify-ios-transport-pool.sh
set -euo pipefail

PORT="${SCARF_VERIFY_PORT:-2222}"
REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$(mktemp -d -t scarf-verify)"

cleanup() {
  if [[ -f "$WORK/sshd.pid" ]]; then kill "$(cat "$WORK/sshd.pid")" 2>/dev/null || true; fi
  rm -rf "$WORK"
}
trap cleanup EXIT

echo "==> ephemeral sshd on 127.0.0.1:$PORT (isolated, torn down on exit)"
ssh-keygen -t ed25519 -f "$WORK/ssh_host_ed25519_key" -N "" -q
chmod 600 "$WORK/ssh_host_ed25519_key"
: > "$WORK/authorized_keys"
chmod 600 "$WORK/authorized_keys"
mkdir -p "$WORK/hermes-home"

cat > "$WORK/sshd_config" <<EOF
Port $PORT
ListenAddress 127.0.0.1
HostKey $WORK/ssh_host_ed25519_key
PidFile $WORK/sshd.pid
AuthorizedKeysFile $WORK/authorized_keys
PasswordAuthentication no
PubkeyAuthentication yes
KbdInteractiveAuthentication no
UsePAM no
StrictModes no
# Aggressive so the UN-pooled churn (fresh handshake per op) deterministically
# trips the limit, proving the pooled+coalesced path is what avoids it.
MaxStartups 2:80:4
MaxSessions 100
Subsystem sftp /usr/libexec/sftp-server
EOF

/usr/sbin/sshd -f "$WORK/sshd_config" -E "$WORK/sshd.log"
sleep 1

echo "==> running CitadelTransportPoolLiveTests"
SCARF_LIVE_SSH_HOST=127.0.0.1 \
SCARF_LIVE_SSH_PORT="$PORT" \
SCARF_LIVE_SSH_USER="$(whoami)" \
SCARF_LIVE_AUTHORIZED_KEYS="$WORK/authorized_keys" \
SCARF_LIVE_HERMES_HOME="$WORK/hermes-home" \
  swift test --package-path "$REPO/scarf/Packages/ScarfIOS" --filter CitadelTransportPoolLiveTests

echo "==> done"
