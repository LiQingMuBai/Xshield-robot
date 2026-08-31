#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Ushield Robot Server-Side Deployment Script
# Runs on Ubuntu target server.
#
# Both ushield-bot1 and ushield-bot2 are INDEPENDENT binaries.
# Flow:
#   1. Build one fresh binary from source
#   2. Stop BOTH ushield-bot1 and ushield-bot2 services
#   3. Backup EXISTING bot1 and bot2 binaries separately (timestamped)
#   4. Replace BOTH bot1 and bot2 with the same newly built binary
#   5. Services remain stopped (per requirement: 暂停服务)
# ============================================================

# ---------- Paths (adjust if server layout changes) ----------
REPO_DIR="/home/ubuntu/ushield/ushield-telegram-bot/new/ushield-robot"
BOT1_DIR="/home/ubuntu/ushield/ushield-telegram-bot/new/ushield-robot"
BOT1_BIN="${BOT1_DIR}/ushield-bot1"
BOT2_DIR="/home/ubuntu/ushield/ushield-telegram-bot/old/ushield-robot"
BOT2_BIN="${BOT2_DIR}/ushield-bot2"
# Backups go to each binary's own directory
BOT1_BACKUP_DIR="${BOT1_DIR}/backups"
BOT2_BACKUP_DIR="${BOT2_DIR}/backups"

BOT1_SERVICE="ushield-bot1"
BOT2_SERVICE="ushield-bot2"

TS="$(date +%Y%m%d_%H%M%S)"
LOG_PREFIX="[deploy-${TS}]"

log()  { echo "${LOG_PREFIX} $*"; }
warn() { echo "${LOG_PREFIX} WARN: $*" >&2; }
die()  { echo "${LOG_PREFIX} ERROR: $*" >&2; exit 1; }

# ---------- Pre-flight checks (Ubuntu-specific) ----------
if [ ! -f /etc/os-release ] || ! grep -qi 'ubuntu' /etc/os-release; then
  warn "this script is designed for Ubuntu; detected OS: $(uname -s)"
fi
command -v systemctl >/dev/null 2>&1 || die "systemctl not found (required for Ubuntu systemd services)"
command -v go        >/dev/null 2>&1 || die "go toolchain not found on PATH"
command -v sha256sum >/dev/null 2>&1 || die "sha256sum not found"
command -v sudo      >/dev/null 2>&1 || die "sudo not found"
[ -d "${REPO_DIR}" ]         || die "repo dir not found: ${REPO_DIR}"
[ -f "${BOT1_BIN}" ]         || warn "current bot1 not found at ${BOT1_BIN}"
[ -f "${BOT2_BIN}" ]         || warn "current bot2 not found at ${BOT2_BIN}"
# Sudo pre-check: ensure user can actually sudo for systemctl (no interactive password mid-script)
if ! sudo -n true 2>/dev/null; then
  warn "sudo may require a password later; systemctl stop/start calls will prompt for it"
fi
mkdir -p "${BOT1_DIR}" "${BOT2_DIR}" "${BOT1_BACKUP_DIR}" "${BOT2_BACKUP_DIR}"

log "===== Ushield deploy started (Ubuntu target, both bot1 + bot2 independent) ====="
log "OS info    = $(grep -E '^(NAME|VERSION)=' /etc/os-release 2>/dev/null | tr '\n' ' ' || uname -s)"
log "hostname   = $(hostname -s 2>/dev/null || echo unknown)"
log "repo       = ${REPO_DIR}"
log "bot1       = ${BOT1_BIN}"
log "bot2       = ${BOT2_BIN}"
log "bot1-bkup  = ${BOT1_BACKUP_DIR}"
log "bot2-bkup  = ${BOT2_BACKUP_DIR}"
log "arch/go    = $(go version 2>/dev/null || echo go-unavailable)"

# ---------- 1. Build ----------
log "Step 1/5: building Linux binary..."
pushd "${REPO_DIR}" >/dev/null
  go mod tidy || die "go mod tidy failed"
  go vet ./... || warn "go vet reported issues (continuing)"
  BUILD_OUT="${REPO_DIR}/ushield-robot_build_${TS}"
  GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "${BUILD_OUT}" ./cmd \
    || die "go build failed"
  chmod +x "${BUILD_OUT}"
  BUILD_SHA="$(sha256sum "${BUILD_OUT}" | cut -d' ' -f1)"
  log "  built => ${BUILD_OUT}  ($(du -h "${BUILD_OUT}" | cut -f1))"
  log "  sha256 = ${BUILD_SHA}"
popd >/dev/null

# ---------- 2. Stop services ----------
log "Step 2/5: stopping systemd services (bot1 + bot2)..."
svc_stop() {
  local s="$1"
  if systemctl is-active --quiet "${s}" 2>/dev/null; then
    log "  stopping ${s}..."
    sudo systemctl stop "${s}" || die "failed to stop ${s}"
    log "  ${s} stopped"
  else
    log "  ${s} already inactive or unknown (skip)"
  fi
}
svc_stop "${BOT1_SERVICE}"
svc_stop "${BOT2_SERVICE}"

# ---------- 3. Backup BOTH existing binaries (to their own directories) ----------
log "Step 3/5: backing up existing bot1 and bot2 to each own directory..."
backup_bin() {
  local src="$1" name="$2" bkdir="$3"
  if [ -f "${src}" ]; then
    local dst="${bkdir}/${name}_${TS}"
    cp -a "${src}" "${dst}"
    log "  backed up ${name} => ${dst}  (sha256: $(sha256sum "${src}" | cut -d' ' -f1))"
  else
    warn "  no existing ${name} found at ${src} (skip backup)"
  fi
}
backup_bin "${BOT1_BIN}" "ushield-bot1" "${BOT1_BACKUP_DIR}"
backup_bin "${BOT2_BIN}" "ushield-bot2" "${BOT2_BACKUP_DIR}"

# ---------- 4. Replace BOTH bot1 and bot2 with the same new binary ----------
log "Step 4/5: installing freshly built binary to BOTH bot1 and bot2..."
install_bin() {
  local dst="$1" name="$2"
  cp -f "${BUILD_OUT}" "${dst}"
  chmod 755 "${dst}"
  log "  installed ${name} => ${dst}  ($(du -h "${dst}" | cut -f1), sha256: $(sha256sum "${dst}" | cut -d' ' -f1))"
}
install_bin "${BOT1_BIN}" "bot1"
install_bin "${BOT2_BIN}" "bot2"

# Verify both installed binaries match the build output
BOT1_SHA="$(sha256sum "${BOT1_BIN}" | cut -d' ' -f1)"
BOT2_SHA="$(sha256sum "${BOT2_BIN}" | cut -d' ' -f1)"
if [ "${BOT1_SHA}" != "${BUILD_SHA}" ] || [ "${BOT2_SHA}" != "${BUILD_SHA}" ]; then
  die "sha256 mismatch after install! build=${BUILD_SHA} bot1=${BOT1_SHA} bot2=${BOT2_SHA}"
fi
log "  sha256 verified: build == bot1 == bot2"

# Cleanup build temp file
rm -f "${BUILD_OUT}"
log "  removed temp build file ${BUILD_OUT}"

# ---------- 5. Services status ----------
log "Step 5/5: confirming services remain stopped..."
report_svc() {
  local s="$1"
  local state
  state="$(systemctl is-active "${s}" 2>/dev/null || echo unknown)"
  log "  ${s}: ${state}"
}
report_svc "${BOT1_SERVICE}"
report_svc "${BOT2_SERVICE}"

log "===== Deploy finished ====="
echo ""
echo "Summary:"
echo "  - Freshly built binary installed to both locations:"
echo "      bot1: ${BOT1_BIN}"
echo "      bot2: ${BOT2_BIN}"
echo "  - Installed sha256: ${BUILD_SHA}"
echo "  - Timestamped backups (each in its own directory):"
echo "      bot1 backup dir: ${BOT1_BACKUP_DIR}"
echo "      bot2 backup dir: ${BOT2_BACKUP_DIR}"
echo "  - Services ${BOT1_SERVICE} / ${BOT2_SERVICE} are STOPPED (per requirement)."
echo ""
echo "To start services manually later:"
echo "  sudo systemctl start ${BOT1_SERVICE}"
echo "  sudo systemctl start ${BOT2_SERVICE}"
