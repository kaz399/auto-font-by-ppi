#!/usr/bin/env bash
set -euo pipefail

###############################################################################
# Default settings
# These can be overridden by command-line options.
###############################################################################

DEFAULT_DRY_RUN=1
DEFAULT_PREFERRED_DISPLAY=""
DEFAULT_APPLY_MODE="full_fonts"

UI_FONT_FAMILY="Cantarell"
DOC_FONT_FAMILY="Cantarell"
MONO_FONT_FAMILY="Monospace"
TITLE_FONT_FAMILY="Cantarell Bold"

# Format:
#   max_ppi:text_scaling:ui_font_size:doc_font_size:mono_font_size:title_font_size
PPI_PROFILES=(
  "110:1.00:11:11:10:11"
  "140:1.10:12:12:11:12"
  "180:1.25:13:13:12:13"
  "240:1.40:14:14:13:14"
  "9999:1.60:16:16:14:16"
)

###############################################################################
# Runtime variables
###############################################################################

DRY_RUN="$DEFAULT_DRY_RUN"
PREFERRED_DISPLAY="$DEFAULT_PREFERRED_DISPLAY"
APPLY_MODE="$DEFAULT_APPLY_MODE"
VERBOSE=0

###############################################################################
# Helpers
###############################################################################

have() {
  command -v "$1" >/dev/null 2>&1
}

log() {
  printf '%s\n' "$*" >&2
}

vlog() {
  if [[ "$VERBOSE" == "1" ]]; then
    printf '%s\n' "$*" >&2
  fi
}

die() {
  log "ERROR: $*"
  exit 1
}

usage() {
  cat <<'EOF'
Usage:
  gnome-auto-font-by-ppi.sh [options]

Purpose:
  Detect the target display, calculate its PPI, select a font profile,
  and update GNOME font settings accordingly.

Display selection:
  1. If --display NAME is given, use that display.
  2. Otherwise, use the primary display if detected.
  3. Otherwise, fall back to the first connected display.

Options:
  --dry-run
      Show what would be changed, but do not apply settings.

  --apply
      Actually apply settings.

  --display NAME
      Use the specified display name instead of auto-detecting.
      Example: --display eDP-1

  --mode MODE
      Set apply mode.
      Allowed values:
        scaling_only
        full_fonts

  --verbose
      Print more diagnostic information.

  --help
      Show this help message and exit.

Examples:
  Dry run with auto-detected primary display:
    gnome-auto-font-by-ppi.sh --dry-run

  Apply settings using auto-detected primary display:
    gnome-auto-font-by-ppi.sh --apply

  Apply settings for eDP-1 explicitly:
    gnome-auto-font-by-ppi.sh --apply --display eDP-1

  Only update text scaling:
    gnome-auto-font-by-ppi.sh --apply --mode scaling_only

Notes:
  - On Xorg, display detection uses xrandr.
  - On GNOME Wayland, it tries Mutter DisplayConfig over D-Bus.
  - Physical monitor size may come from EDID and can be inaccurate on some setups.
EOF
}

calc_ppi() {
  local wpx="$1" hpx="$2" wmm="$3" hmm="$4"
  awk -v wpx="$wpx" -v hpx="$hpx" -v wmm="$wmm" -v hmm="$hmm" '
    BEGIN {
      if (wmm <= 0 || hmm <= 0) exit 1
      diag_px = sqrt((wpx*wpx) + (hpx*hpx))
      diag_in = sqrt(((wmm/25.4)*(wmm/25.4)) + ((hmm/25.4)*(hmm/25.4)))
      printf "%.2f\n", diag_px / diag_in
    }'
}

###############################################################################
# Argument parsing
###############################################################################

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --dry-run)
        DRY_RUN=1
        shift
        ;;
      --apply)
        DRY_RUN=0
        shift
        ;;
      --display)
        [[ $# -ge 2 ]] || die "--display requires a value"
        PREFERRED_DISPLAY="$2"
        shift 2
        ;;
      --mode)
        [[ $# -ge 2 ]] || die "--mode requires a value"
        case "$2" in
          scaling_only|full_fonts)
            APPLY_MODE="$2"
            ;;
          *)
            die "Invalid --mode value: $2"
            ;;
        esac
        shift 2
        ;;
      --verbose)
        VERBOSE=1
        shift
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      *)
        die "Unknown option: $1 (use --help)"
        ;;
    esac
  done
}

###############################################################################
# Display info collection
# Output format:
#   <name>\t<is_primary>\t<wpx>\t<hpx>\t<wmm>\t<hmm>\t<ppi>
###############################################################################

from_xrandr() {
  have xrandr || return 1

  xrandr --query | awk '
    / connected/ {
      name=$1
      primary=0
      wpx=""; hpx=""; wmm=""; hmm=""

      for (i=1; i<=NF; i++) {
        if ($i == "primary") {
          primary=1
        }
        if ($i ~ /^[0-9]+x[0-9]+\+[0-9]+\+[0-9]+$/) {
          split($i, a, "+")
          split(a[1], b, "x")
          wpx=b[1]
          hpx=b[2]
        }
        if ($i ~ /^[0-9]+mm$/ && (i+2)<=NF && $(i+1)=="x" && $(i+2) ~ /^[0-9]+mm$/) {
          wmm=$i
          hmm=$(i+2)
          sub(/mm$/, "", wmm)
          sub(/mm$/, "", hmm)
        }
      }

      if (wpx != "" && hpx != "" && wmm != "" && hmm != "") {
        printf "%s\t%s\t%s\t%s\t%s\t%s\n", name, primary, wpx, hpx, wmm, hmm
      }
    }
  ' | while IFS=$'\t' read -r name primary wpx hpx wmm hmm; do
    ppi="$(calc_ppi "$wpx" "$hpx" "$wmm" "$hmm")" || continue
    printf "%s\t%s\t%s\t%s\t%s\t%s\t%s\n" "$name" "$primary" "$wpx" "$hpx" "$wmm" "$hmm" "$ppi"
  done
}

from_gnome_wayland() {
  have gdbus || return 1
  have python3 || return 1

  local out
  out="$(gdbus call --session \
    --dest org.gnome.Mutter.DisplayConfig \
    --object-path /org/gnome/Mutter/DisplayConfig \
    --method org.gnome.Mutter.DisplayConfig.GetCurrentState 2>/dev/null)" || return 1

  python3 - <<'PY' "$out"
import re
import sys
import math

s = sys.argv[1]

conn_re = re.compile(r"'((?:HDMI|DP|eDP|DVI|VGA|Virtual|DisplayPort)[^']*)'")
mode_re = re.compile(r'[,(\[]\s*(\d{3,5})\s*,\s*(\d{3,5})\s*,\s*(?:[0-9]+(?:\.[0-9]+)?)')
mm_re = re.compile(r'[,(\[]\s*(\d{2,5})\s*,\s*(\d{2,5})\s*[,)\]]')
primary_hint_re = re.compile(r'primary', re.IGNORECASE)

seen = set()

for m in conn_re.finditer(s):
    name = m.group(1)
    tail = s[m.end():m.end()+2500]

    mode = mode_re.search(tail)
    mm = mm_re.search(tail)
    if not mode or not mm:
        continue

    wpx, hpx = map(int, mode.groups())
    wmm, hmm = map(int, mm.groups())
    if wmm <= 0 or hmm <= 0:
        continue

    ppi = math.hypot(wpx, hpx) / math.hypot(wmm / 25.4, hmm / 25.4)
    is_primary = 1 if primary_hint_re.search(tail[:500]) else 0

    key = (name, is_primary, wpx, hpx, wmm, hmm)
    if key in seen:
        continue
    seen.add(key)

    print(f"{name}\t{is_primary}\t{wpx}\t{hpx}\t{wmm}\t{hmm}\t{ppi:.2f}")
PY
}

get_display_info() {
  local session_type="${XDG_SESSION_TYPE:-}"
  local desktop="${XDG_CURRENT_DESKTOP:-}"

  vlog "Session type: ${session_type:-unknown}"
  vlog "Desktop: ${desktop:-unknown}"

  case "$session_type" in
    x11|xorg)
      from_xrandr && return 0
      ;;
    wayland)
      if [[ "$desktop" == *GNOME* ]]; then
        from_gnome_wayland && return 0
      fi
      from_xrandr && return 0
      ;;
  esac

  from_xrandr && return 0
  from_gnome_wayland && return 0
  return 1
}

###############################################################################
# Display selection
###############################################################################

get_display_line_by_name() {
  local info="$1"
  local target="$2"

  printf '%s\n' "$info" | awk -F '\t' -v target="$target" '
    $1 == target { print; found=1; exit }
    END { if (!found) exit 1 }
  '
}

get_primary_display_line() {
  local info="$1"

  local primary_line
  primary_line="$(printf '%s\n' "$info" | awk -F '\t' '$2 == 1 { print; exit }')"
  if [[ -n "$primary_line" ]]; then
    printf '%s\n' "$primary_line"
    return 0
  fi

  printf '%s\n' "$info" | head -n 1
}

select_target_display_line() {
  local info="$1"

  if [[ -n "$PREFERRED_DISPLAY" ]]; then
    get_display_line_by_name "$info" "$PREFERRED_DISPLAY" && return 0
    die "Preferred display '${PREFERRED_DISPLAY}' was not found."
  fi

  get_primary_display_line "$info"
}

###############################################################################
# Profile selection
###############################################################################

select_profile() {
  local detected_ppi="$1"

  for profile in "${PPI_PROFILES[@]}"; do
    IFS=':' read -r max_ppi scaling ui_size doc_size mono_size title_size <<<"$profile"
    if awk -v ppi="$detected_ppi" -v limit="$max_ppi" 'BEGIN { exit !(ppi <= limit) }'; then
      printf '%s\n' "$profile"
      return 0
    fi
  done

  return 1
}

###############################################################################
# Apply GNOME settings
###############################################################################

apply_scaling_only() {
  local scaling="$1"

  if [[ "$DRY_RUN" == "1" ]]; then
    log "[DRY-RUN] gsettings set org.gnome.desktop.interface text-scaling-factor $scaling"
    return 0
  fi

  gsettings set org.gnome.desktop.interface text-scaling-factor "$scaling"
}

apply_full_fonts() {
  local scaling="$1"
  local ui_size="$2"
  local doc_size="$3"
  local mono_size="$4"
  local title_size="$5"

  if [[ "$DRY_RUN" == "1" ]]; then
    log "[DRY-RUN] gsettings set org.gnome.desktop.interface text-scaling-factor $scaling"
    log "[DRY-RUN] gsettings set org.gnome.desktop.interface font-name '${UI_FONT_FAMILY} ${ui_size}'"
    log "[DRY-RUN] gsettings set org.gnome.desktop.interface document-font-name '${DOC_FONT_FAMILY} ${doc_size}'"
    log "[DRY-RUN] gsettings set org.gnome.desktop.interface monospace-font-name '${MONO_FONT_FAMILY} ${mono_size}'"
    log "[DRY-RUN] gsettings set org.gnome.desktop.wm.preferences titlebar-font '${TITLE_FONT_FAMILY} ${title_size}'"
    return 0
  fi

  gsettings set org.gnome.desktop.interface text-scaling-factor "$scaling"
  gsettings set org.gnome.desktop.interface font-name "${UI_FONT_FAMILY} ${ui_size}"
  gsettings set org.gnome.desktop.interface document-font-name "${DOC_FONT_FAMILY} ${doc_size}"
  gsettings set org.gnome.desktop.interface monospace-font-name "${MONO_FONT_FAMILY} ${mono_size}"
  gsettings set org.gnome.desktop.wm.preferences titlebar-font "${TITLE_FONT_FAMILY} ${title_size}"
}

###############################################################################
# Main
###############################################################################

main() {
  parse_args "$@"

  have gsettings || die "gsettings not found."

  local info
  info="$(get_display_info)" || die "Could not determine display information."
  [[ -n "$info" ]] || die "No connected displays with usable size information were found."

  log "Detected displays:"
  printf '%s\n' "$info" | awk -F '\t' '
    {
      mark = ($2 == 1 ? " [primary]" : "")
      printf "  %s%s: %sx%s px, %sx%s mm, PPI=%s\n", $1, mark, $3, $4, $5, $6, $7
    }' >&2

  local target_line
  target_line="$(select_target_display_line "$info")" || die "Could not select target display."

  IFS=$'\t' read -r name is_primary wpx hpx wmm hmm ppi <<<"$target_line"

  if [[ -n "$PREFERRED_DISPLAY" ]]; then
    log "Selected display by override: $name"
  else
    log "Selected display by auto-detection: $name"
  fi

  log "Target display details: primary=$is_primary, resolution=${wpx}x${hpx}, size=${wmm}x${hmm} mm, PPI=$ppi"

  local profile
  profile="$(select_profile "$ppi")" || die "No matching profile found for PPI=$ppi"

  IFS=':' read -r max_ppi_limit scaling ui_size doc_size mono_size title_size <<<"$profile"

  log "Selected profile:"
  log "  max_ppi <= $max_ppi_limit"
  log "  text_scaling_factor = $scaling"
  log "  ui_font_size        = $ui_size"
  log "  document_font_size  = $doc_size"
  log "  monospace_font_size = $mono_size"
  log "  titlebar_font_size  = $title_size"

  case "$APPLY_MODE" in
    scaling_only)
      apply_scaling_only "$scaling"
      ;;
    full_fonts)
      apply_full_fonts "$scaling" "$ui_size" "$doc_size" "$mono_size" "$title_size"
      ;;
    *)
      die "Invalid APPLY_MODE: $APPLY_MODE"
      ;;
  esac

  if [[ "$DRY_RUN" == "1" ]]; then
    log "Dry run completed. No settings were changed."
  else
    log "GNOME font settings updated."
  fi
}

main "$@"
