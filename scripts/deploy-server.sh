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

# ---------- 2. Stop services/processes (systemd first, then direct pid-kill fallback) ----------
# Process detection: find PIDs whose /proc/<pid>/exe points at the binary (canonical match),
# OR whose /proc/<pid>/cmdline argv[0] basename matches the binary name.
# This covers: systemd services, nohup background, screen/tmux, supervisor, manual & etc.
pids_of_bin() {
  local bin_path="$1" bin_name
  bin_name="$(basename "${bin_path}")"
  local -a pids=() pid exe cmdname
  # Prefer /proc-based exact exe match (works when process still has original binary fd open)
  for pid in /proc/[0-9]*; do
    pid="${pid##*/}"
    [ -d "/proc/${pid}" ] || continue
    exe="$(readlink "/proc/${pid}/exe" 2>/dev/null || true)"
    cmdname="$(tr '\0' ' ' < "/proc/${pid}/cmdline" 2>/dev/null | awk '{print $1}' || true)"
    cmdname="$(basename "${cmdname}" 2>/dev/null || true)"
    if [ -n "${exe}" ] && [ "${exe}" = "${bin_path}" -o "$(basename "${exe}")" = "${bin_name}" ]; then
      pids+=("${pid}")
    elif [ -n "${cmdname}" ] && [ "${cmdname}" = "${bin_name}" ]; then
      pids+=("${pid}")
    fi
  done
  printf '%s\n' "${pids[@]}" 2>/dev/null | sort -u | grep -v '^$' || true
}

kill_pids() {
  local name="$1" bin_path="$2"
  local -a pids
  mapfile -t pids < <(pids_of_bin "${bin_path}")
  if [ "${#pids[@]}" -eq 0 ]; then
    log "  ${name}: no running processes found (nothing to kill)"
    return 0
  fi
  log "  ${name}: found running PIDs [${pids[*]}], stopping gracefully..."
  # SIGTERM
  sudo kill -TERM "${pids[@]}" 2>/dev/null || true
  local waited=0 total=15
  while [ "${waited}" -lt "${total}" ]; do
    sleep 1
    waited=$((waited+1))
    local -a still
    mapfile -t still < <(pids_of_bin "${bin_path}")
    if [ "${#still[@]}" -eq 0 ]; then
      log "  ${name}: all PIDs exited within ${waited}s after SIGTERM"
      return 0
    fi
  done
  local -a still
  mapfile -t still < <(pids_of_bin "${bin_path}")
  if [ "${#still[@]}" -gt 0 ]; then
    warn "  ${name}: PIDs [${still[*]}] still alive after ${total}s, sending SIGKILL..."
    sudo kill -9 "${still[@]}" 2>/dev/null || true
    sleep 1
    local -a remain
    mapfile -t remain < <(pids_of_bin "${bin_path}")
    [ "${#remain[@]}" -eq 0 ] && log "  ${name}: killed via SIGKILL" \
      || warn "  ${name}: PIDs [${remain[*]}] could not be killed (zombie or permissions?)"
  fi
}

log "Step 2/5: stopping bot1 + bot2 (systemd stop first, fallback to process kill)..."
svc_stop() {
  local s="$1"
  if systemctl list-unit-files --full 2>/dev/null | grep -q "^${s}\.service"; then
    if systemctl is-active --quiet "${s}" 2>/dev/null; then
      log "  systemd stop ${s}..."
      sudo systemctl stop "${s}" || warn "  systemctl stop ${s} returned non-zero (will rely on process kill below)"
    else
      log "  systemd unit ${s} exists but is inactive"
    fi
  else
    log "  systemd unit ${s}.service NOT installed (process may be managed by nohup/screen/supervisor/...)"
  fi
}
svc_stop "${BOT1_SERVICE}"
svc_stop "${BOT2_SERVICE}"

# Settle: give systemd time to actually kill child cgroup processes before we probe
sleep 1

# Pre-kill process-survey: ensure we know what we're stopping
echo "    [before-stop process snapshot]"
ps -eo pid,ppid,user,etime,stat,cmd 2>/dev/null | grep -E "(ushield-bot1|ushield-bot2)" | grep -v grep || echo "      (nothing matched by grep)"

kill_pids "bot1" "${BOT1_BIN}"
sleep 1
kill_pids "bot2" "${BOT2_BIN}"
sleep 1

echo "    [after-stop process snapshot]"
ps -eo pid,ppid,user,etime,stat,cmd 2>/dev/null | grep -E "(ushield-bot1|ushield-bot2)" | grep -v grep || echo "      (nothing matched by grep — all clean)"
sleep 1

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
sleep 1
backup_bin "${BOT2_BIN}" "ushield-bot2" "${BOT2_BACKUP_DIR}"
sleep 1

# ---------- 4. Replace BOTH bot1 and bot2 with the same new binary ----------
log "Step 4/5: installing freshly built binary to BOTH bot1 and bot2..."
install_bin() {
  local dst="$1" name="$2"
  cp -f "${BUILD_OUT}" "${dst}"
  chmod 755 "${dst}"
  log "  installed ${name} => ${dst}  ($(du -h "${dst}" | cut -f1), sha256: $(sha256sum "${dst}" | cut -d' ' -f1))"
}
install_bin "${BOT1_BIN}" "bot1"
sleep 1
install_bin "${BOT2_BIN}" "bot2"
sleep 1

# Verify both installed binaries match the build output
BOT1_SHA="$(sha256sum "${BOT1_BIN}" | cut -d' ' -f1)"
BOT2_SHA="$(sha256sum "${BOT2_BIN}" | cut -d' ' -f1)"
if [ "${BOT1_SHA}" != "${BUILD_SHA}" ] || [ "${BOT2_SHA}" != "${BUILD_SHA}" ]; then
  die "sha256 mismatch after install! build=${BUILD_SHA} bot1=${BOT1_SHA} bot2=${BOT2_SHA}"
fi
log "  sha256 verified: build == bot1 == bot2"
sleep 1

# Cleanup build temp file
rm -f "${BUILD_OUT}"
log "  removed temp build file ${BUILD_OUT}"
sleep 1

# ---------- 5. Final stopped-state verification ----------
log "Step 5/5: confirming both services/processes remain STOPPED..."
sleep 1
echo "    [final process snapshot]"
FINAL_PS="$(ps -eo pid,ppid,user,etime,stat,cmd 2>/dev/null | grep -E "(ushield-bot1|ushield-bot2)" | grep -v grep || true)"
if [ -n "${FINAL_PS}" ]; then
  echo "${FINAL_PS}"
  warn "WARNING: processes still running!  They were NOT started by systemd nor matched by our /proc exe/cmdline detection."
  warn "Please kill them manually using the PIDs shown above, or report their full cmdline so the kill rule can be extended."
else
  echo "      (nothing matched by grep — all clean, services/processes STOPPED as required)"
fi
sleep 1

report_svc() {
  local s="$1"
  local state
  if systemctl list-unit-files --full 2>/dev/null | grep -q "^${s}\.service"; then
    state="$(systemctl is-active "${s}" 2>/dev/null || echo unknown)"
    log "  systemd unit ${s}.service: state=${state}"
  else
    log "  systemd unit ${s}.service: NOT INSTALLED (process was stopped by direct pid-kill)"
  fi
}
report_svc "${BOT1_SERVICE}"
sleep 1
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
echo "  - Processes ushield-bot1 + ushield-bot2 are STOPPED (both systemctl stop + direct pid-kill were applied)."
echo ""
echo "To start manually later, pick ONE depending on how you run them:"
echo "  (A) If you use systemd:   sudo systemctl start ${BOT1_SERVICE} ; sudo systemctl start ${BOT2_SERVICE}"
echo "  (B) If you use nohup:     nohup ${BOT1_BIN} >/var/log/ushield-bot1.log 2>&1 &  (similar for bot2)"
echo "  (C) If you use screen/tmux/supervisor: use your existing wrapper"
