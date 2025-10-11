#!/usr/bin/env bash
###############################################################################
# verify_quick.sh - Quick verification for P2 (Carnet 201905884)
# Corregido para usar espacios en lugar de = en parámetros
###############################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Config
API_BASE="http://localhost:8080"
TEST_DIR="/tmp/godisk_verify_84"
DISK_PATH="${TEST_DIR}/test_disk.mia"
PASSED=0
FAILED=0

log_pass() { echo -e "${GREEN}[PASS]${NC} $*"; ((PASSED++)); }
log_fail() { echo -e "${RED}[FAIL]${NC} $*"; ((FAILED++)); }
log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_section() { echo ""; echo -e "${BLUE}========== $* ==========${NC}"; }

run_cmd() {
    local cmd="$1"
    curl -s -X POST "${API_BASE}/api/cmd/run" \
        -H "Content-Type: application/json" \
        -d "{\"line\":\"$cmd\"}"
}

check_ok() {
    local resp="$1"
    echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1
}

# Setup
mkdir -p "$TEST_DIR"
rm -f "$DISK_PATH"

echo ""
echo "╔══════════════════════════════════════════════╗"
echo "║  Quick Verifier P2 - Carnet 84              ║"
echo "╚══════════════════════════════════════════════╝"

# P1: MKDISK
log_section "P1: MKDISK"
resp=$(run_cmd "mkdisk -path $DISK_PATH -size 5 -unit m")
if check_ok "$resp" && [[ -f "$DISK_PATH" ]]; then
    log_pass "MKDISK creó disco de 5MB"
else
    log_fail "MKDISK falló: $(echo "$resp" | jq -r '.error')"
fi

# P1: FDISK - Primaria
log_section "P1: FDISK Primaria"
resp=$(run_cmd "fdisk -path $DISK_PATH -mode add -name Part1 -size 2 -unit m -type p")
if check_ok "$resp"; then
    log_pass "FDISK agregó partición primaria Part1"
else
    log_fail "FDISK primaria falló: $(echo "$resp" | jq -r '.error')"
fi

# P1: FDISK - Extendida
log_section "P1: FDISK Extendida"
resp=$(run_cmd "fdisk -path $DISK_PATH -mode add -name Extended -size 2 -unit m -type e")
if check_ok "$resp"; then
    log_pass "FDISK agregó partición extendida"
else
    log_fail "FDISK extendida falló: $(echo "$resp" | jq -r '.error')"
fi

# P1: FDISK - Lógica
log_section "P1: FDISK Lógica (EBR)"
resp=$(run_cmd "fdisk -path $DISK_PATH -mode add -name Logic1 -size 1 -unit m -type l")
if check_ok "$resp"; then
    log_pass "FDISK agregó partición lógica (EBR)"
else
    log_fail "FDISK lógica falló: $(echo "$resp" | jq -r '.error')"
fi

# P1: MOUNT
log_section "P1: MOUNT con ID formato 84X"
resp=$(run_cmd "mount -path $DISK_PATH -name Part1")
if check_ok "$resp"; then
    MOUNT_ID=$(echo "$resp" | jq -r '.output' | grep -oP '841[A-Z]' | head -1)
    if [[ "$MOUNT_ID" =~ ^84[0-9]+[A-Z]$ ]]; then
        log_pass "MOUNT asignó ID correcto: $MOUNT_ID"
    else
        log_fail "MOUNT ID no sigue formato 84X: $MOUNT_ID"
    fi
else
    log_fail "MOUNT falló: $(echo "$resp" | jq -r '.error')"
    MOUNT_ID=""
fi

# P1: MOUNTED
log_section "P1: MOUNTED"
resp=$(run_cmd "mounted")
if check_ok "$resp"; then
    log_pass "MOUNTED lista particiones"
    echo "$resp" | jq -r '.output'
else
    log_fail "MOUNTED falló"
fi

# P1: MKFS EXT2
if [[ -n "$MOUNT_ID" ]]; then
    log_section "P1: MKFS EXT2"
    resp=$(run_cmd "mkfs -id $MOUNT_ID -fs 2fs")
    if check_ok "$resp"; then
        log_pass "MKFS EXT2 formateó partición"
    else
        log_fail "MKFS EXT2 falló: $(echo "$resp" | jq -r '.error')"
    fi
fi

# P2: MOUNT segunda partición para EXT3
log_section "P2: MOUNT para EXT3"
resp=$(run_cmd "mount -path $DISK_PATH -name Logic1")
if check_ok "$resp"; then
    EXT3_ID=$(echo "$resp" | jq -r '.output' | grep -oP '84[0-9]+[A-Z]' | head -1)
    log_pass "Partición Logic1 montada para EXT3: $EXT3_ID"
else
    log_fail "MOUNT Logic1 falló: $(echo "$resp" | jq -r '.error')"
    EXT3_ID=""
fi

# P2: MKFS EXT3
if [[ -n "$EXT3_ID" ]]; then
    log_section "P2: MKFS EXT3 con Journal"
    resp=$(run_cmd "mkfs -id $EXT3_ID -fs 3fs")
    if check_ok "$resp"; then
        log_pass "MKFS EXT3 creó filesystem con journal"
    else
        log_fail "MKFS EXT3 falló: $(echo "$resp" | jq -r '.error')"
    fi
fi

# P2: JOURNALING
if [[ -n "$EXT3_ID" ]]; then
    log_section "P2: Comando JOURNALING"
    resp=$(run_cmd "journaling -id $EXT3_ID")
    if check_ok "$resp"; then
        log_pass "JOURNALING devolvió entradas del journal"
    else
        log_fail "JOURNALING falló: $(echo "$resp" | jq -r '.error')"
    fi
fi

# P2: RECOVERY
if [[ -n "$EXT3_ID" ]]; then
    log_section "P2: Comando RECOVERY"
    resp=$(run_cmd "recovery -id $EXT3_ID")
    if check_ok "$resp"; then
        log_pass "RECOVERY ejecutó sin errores"
    else
        log_fail "RECOVERY falló: $(echo "$resp" | jq -r '.error')"
    fi
fi

# P2: LOSS
if [[ -n "$EXT3_ID" ]]; then
    log_section "P2: Comando LOSS"
    resp=$(run_cmd "loss -id $EXT3_ID")
    if check_ok "$resp"; then
        log_pass "LOSS ejecutó sin errores"
    else
        log_fail "LOSS falló: $(echo "$resp" | jq -r '.error')"
    fi
fi

# P2: Reportes DOT
if [[ -n "$MOUNT_ID" ]]; then
    log_section "P2: Reportes DOT"

    # MBR Report
    mbr=$(curl -s "http://localhost:8080/api/reports/mbr?id=$MOUNT_ID")
    if echo "$mbr" | grep -q "digraph MBR"; then
        log_pass "Reporte MBR devuelve formato DOT válido"
    else
        log_fail "Reporte MBR no es DOT válido"
    fi

    # DISK Report
    disk=$(curl -s "http://localhost:8080/api/reports/disk?id=$MOUNT_ID")
    if echo "$disk" | grep -q "digraph"; then
        log_pass "Reporte DISK devuelve formato DOT válido"
    else
        log_fail "Reporte DISK no es DOT válido"
    fi

    # TREE Report
    tree=$(curl -s "http://localhost:8080/api/reports/tree?id=$MOUNT_ID")
    if echo "$tree" | grep -q "digraph"; then
        log_pass "Reporte TREE devuelve formato DOT válido"
    else
        log_fail "Reporte TREE no es DOT válido"
    fi

    # SB Report
    sb=$(curl -s "http://localhost:8080/api/reports/sb?id=$MOUNT_ID")
    if echo "$sb" | grep -q "digraph"; then
        log_pass "Reporte SUPERBLOCK devuelve formato DOT válido"
    else
        log_fail "Reporte SUPERBLOCK no es DOT válido"
    fi
fi

# P2: Journal Report
if [[ -n "$EXT3_ID" ]]; then
    journal=$(curl -s "http://localhost:8080/api/reports/journal?id=$EXT3_ID")
    if echo "$journal" | grep -q "digraph"; then
        log_pass "Reporte JOURNAL devuelve formato DOT válido"
    else
        log_fail "Reporte JOURNAL no es DOT válido"
    fi
fi

# P2: HTTP Endpoints EXT3
if [[ -n "$EXT3_ID" ]]; then
    log_section "P2: Endpoints HTTP EXT3"

    # GET /api/ext3/journal
    resp=$(curl -s "http://localhost:8080/api/ext3/journal?id=$EXT3_ID")
    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_pass "GET /api/ext3/journal responde correctamente"
    else
        log_fail "GET /api/ext3/journal falló"
    fi

    # POST /api/ext3/recovery
    resp=$(curl -s -X POST "http://localhost:8080/api/ext3/recovery" \
        -H "Content-Type: application/json" \
        -d "{\"id\":\"$EXT3_ID\"}")
    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_pass "POST /api/ext3/recovery responde correctamente"
    else
        log_fail "POST /api/ext3/recovery falló"
    fi

    # POST /api/ext3/loss
    resp=$(curl -s -X POST "http://localhost:8080/api/ext3/loss" \
        -H "Content-Type: application/json" \
        -d "{\"id\":\"$EXT3_ID\"}")
    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_pass "POST /api/ext3/loss responde correctamente"
    else
        log_fail "POST /api/ext3/loss falló"
    fi
fi

# P2: API Endpoints
log_section "P2: API Endpoints Generales"

# /api/commands
resp=$(curl -s "http://localhost:8080/api/commands")
if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
    log_pass "GET /api/commands lista comandos"
else
    log_fail "GET /api/commands falló"
fi

# /api/disks/mounted
resp=$(curl -s "http://localhost:8080/api/disks/mounted")
if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
    log_pass "GET /api/disks/mounted funciona"
else
    log_fail "GET /api/disks/mounted falló"
fi

# /api/disks/list
resp=$(curl -s "http://localhost:8080/api/disks/list?path=$TEST_DIR")
if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
    log_pass "GET /api/disks/list funciona"
else
    log_fail "GET /api/disks/list falló"
fi

# Cleanup
log_section "Limpieza"
if [[ -n "$MOUNT_ID" ]]; then
    run_cmd "unmount -id $MOUNT_ID" >/dev/null 2>&1
fi
if [[ -n "$EXT3_ID" ]]; then
    run_cmd "unmount -id $EXT3_ID" >/dev/null 2>&1
fi
rm -rf "$TEST_DIR"
log_pass "Limpieza completada"

# Summary
echo ""
log_section "RESUMEN"
TOTAL=$((PASSED + FAILED))
RATE=$((PASSED * 100 / TOTAL))
echo "Total tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "Success rate: ${RATE}%"
echo ""

if [[ $FAILED -eq 0 ]]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED!${NC}"
    echo "Tu P2 (Carnet 84) está completo y funcional."
    exit 0
else
    echo -e "${RED}✗ SOME TESTS FAILED${NC}"
    echo "Revisa los mensajes arriba para identificar problemas."
    exit 1
fi
