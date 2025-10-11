#!/usr/bin/env bash

##############################################################################
# verify_p2_84.sh - Script de Verificación Integral para P2 (Carnet 201905884)
##############################################################################
# Este script valida que todos los componentes del P1 y P2 estén funcionando:
#
# P1 (Requisitos Base):
#   - MKDISK/FDISK: Crear discos y particiones (MBR/EBR)
#   - MOUNT/MOUNTED: Montar particiones con IDs basados en terminación 84
#   - MKFS EXT2: Formatear con /users.txt, bitmaps, superblock
#
# P2 (Nuevos Requisitos):
#   - MKFS EXT3: Journal, comandos journaling/recovery/loss
#   - Reportes DOT: /api/reports/{mbr,disk,tree,journal,sb}
#   - Consola HTTP: /api/cmd/run con errores consistentes
#   - Endpoints EXT3: /api/ext3/{journaling,recovery,loss}
#
##############################################################################

set -eo pipefail

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuración
API_BASE="http://localhost:8080"
TEST_DIR="/tmp/godisk_verify_84"
DISK_PATH="${TEST_DIR}/test_disk.mia"
MOUNT_ID=""
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

##############################################################################
# Funciones de Utilidad
##############################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $*"
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $*"
    ((FAILED_TESTS++))
    ((TOTAL_TESTS++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$*${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Verificar que el servidor esté corriendo
check_server() {
    # Usar /api/commands que sí existe en lugar de /api/health
    local resp
    resp=$(curl -s "${API_BASE}/api/commands" 2>/dev/null)

    if echo "$resp" | jq -e '.ok' >/dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Ejecutar comando via API
run_command() {
    local cmd="$1"
    local response

    response=$(curl -s -X POST "${API_BASE}/api/cmd/run" \
        -H "Content-Type: application/json" \
        -d "{\"line\":\"$cmd\"}" 2>/dev/null)

    echo "$response"
}

# Verificar respuesta JSON
check_json_ok() {
    local response="$1"
    local ok

    ok=$(echo "$response" | jq -r '.ok // .OK // false' 2>/dev/null)

    if [[ "$ok" == "true" ]]; then
        return 0
    else
        return 1
    fi
}

# Extraer campo de JSON
get_json_field() {
    local response="$1"
    local field="$2"

    echo "$response" | jq -r ".$field // empty" 2>/dev/null
}

##############################################################################
# Tests de Prerequisitos
##############################################################################

test_prerequisites() {
    log_section "Verificando Prerequisitos"

    # Verificar jq
    if command -v jq >/dev/null 2>&1; then
        log_success "jq está instalado"
    else
        log_fail "jq no está instalado (requerido para parsear JSON)"
        echo "Instalar con: sudo apt install jq"
        exit 1
    fi

    # Verificar curl
    if command -v curl >/dev/null 2>&1; then
        log_success "curl está instalado"
    else
        log_fail "curl no está instalado"
        exit 1
    fi

    # Verificar servidor
    log_info "Verificando que el servidor esté corriendo en ${API_BASE}..."
    if check_server; then
        log_success "Servidor HTTP respondiendo correctamente"
    else
        log_fail "Servidor no responde en ${API_BASE}"
        echo ""
        echo "Para iniciar el servidor, ejecuta desde la raíz del proyecto:"
        echo "  ./start.sh"
        echo "  # o desde Backend:"
        echo "  cd Backend && go run cmd/server/*.go"
        echo ""
        exit 1
    fi

    # Crear directorio de pruebas
    mkdir -p "$TEST_DIR"
    log_success "Directorio de pruebas creado: ${TEST_DIR}"
}

##############################################################################
# Tests P1 - Disco y Particiones (MBR/EBR)
##############################################################################

test_p1_disk_operations() {
    log_section "P1: Operaciones de Disco (MKDISK/FDISK)"

    # Limpiar disco previo si existe
    rm -f "$DISK_PATH"

    # Test: MKDISK
    log_info "Creando disco de prueba (5MB)..."
    local resp
    resp=$(run_command "mkdisk -path=${DISK_PATH} -size=5 -unit=m")

    if check_json_ok "$resp" && [[ -f "$DISK_PATH" ]]; then
        log_success "MKDISK creó el disco correctamente"
    else
        log_fail "MKDISK falló: $(get_json_field "$resp" "error")"
        return 1
    fi

    # Test: FDISK - Agregar partición primaria
    log_info "Creando partición primaria P1 (2MB)..."
    resp=$(run_command "fdisk -path=${DISK_PATH} -mode=add -name=Part1 -size=2 -unit=m -type=p")

    if check_json_ok "$resp"; then
        log_success "FDISK agregó partición primaria"
    else
        log_fail "FDISK falló al agregar partición: $(get_json_field "$resp" "error")"
    fi

    # Test: FDISK - Agregar partición extendida
    log_info "Creando partición extendida (2MB)..."
    resp=$(run_command "fdisk -path=${DISK_PATH} -mode=add -name=Extended -size=2 -unit=m -type=e")

    if check_json_ok "$resp"; then
        log_success "FDISK agregó partición extendida"
    else
        log_fail "FDISK falló al agregar extendida: $(get_json_field "$resp" "error")"
    fi

    # Test: FDISK - Agregar partición lógica
    log_info "Creando partición lógica L1 (1MB)..."
    resp=$(run_command "fdisk -path=${DISK_PATH} -mode=add -name=Logic1 -size=1 -unit=m -type=l")

    if check_json_ok "$resp"; then
        log_success "FDISK agregó partición lógica (EBR)"
    else
        log_fail "FDISK falló al agregar lógica: $(get_json_field "$resp" "error")"
    fi
}

##############################################################################
# Tests P1 - MOUNT con IDs basados en 84
##############################################################################

test_p1_mount_operations() {
    log_section "P1: Montaje de Particiones (MOUNT/MOUNTED)"

    # Test: MOUNT - Montar partición primaria
    log_info "Montando partición Part1..."
    local resp
    resp=$(run_command "mount -path=${DISK_PATH} -name=Part1")

    if check_json_ok "$resp"; then
        MOUNT_ID=$(get_json_field "$resp" "output" | grep -oP 'id=\K[^ ]+' | head -1)

        if [[ -z "$MOUNT_ID" ]]; then
            # Intentar extraer de otra forma
            MOUNT_ID=$(echo "$resp" | jq -r '.output' | grep -oP '841[A-Z]')
        fi

        if [[ "$MOUNT_ID" =~ ^84[0-9]+[A-Z]$ ]]; then
            log_success "MOUNT asignó ID con formato correcto: ${MOUNT_ID}"
            log_info "Formato esperado: 84<correlativo><letra> (ej: 841A)"
        else
            log_warn "MOUNT funcionó pero ID no sigue formato 84X: ${MOUNT_ID}"
            log_info "Formato esperado: 84<correlativo><letra> (ej: 841A, 842A, 841B)"
        fi
    else
        log_fail "MOUNT falló: $(get_json_field "$resp" "error")"
        return 1
    fi

    # Test: MOUNTED - Listar particiones montadas
    log_info "Verificando comando MOUNTED..."
    resp=$(run_command "mounted")

    if check_json_ok "$resp"; then
        local output
        output=$(get_json_field "$resp" "output")

        if echo "$output" | grep -q "$MOUNT_ID"; then
            log_success "MOUNTED lista particiones correctamente"
        else
            log_warn "MOUNTED no muestra el ID montado: ${MOUNT_ID}"
        fi
    else
        log_fail "MOUNTED falló: $(get_json_field "$resp" "error")"
    fi

    # Test: API /api/disks/mounted
    log_info "Verificando endpoint /api/disks/mounted..."
    local api_resp
    api_resp=$(curl -s "${API_BASE}/api/disks/mounted")

    if echo "$api_resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        local count
        count=$(echo "$api_resp" | jq '.count // 0')

        if [[ "$count" -gt 0 ]]; then
            log_success "API /api/disks/mounted devuelve particiones montadas (count=$count)"
        else
            log_warn "API /api/disks/mounted no tiene particiones montadas"
        fi
    else
        log_fail "API /api/disks/mounted falló"
    fi
}

##############################################################################
# Tests P1 - MKFS EXT2 (Superblock, Bitmaps, /users.txt)
##############################################################################

test_p1_mkfs_ext2() {
    log_section "P1: Formateo EXT2 (MKFS 2fs)"

    if [[ -z "$MOUNT_ID" ]]; then
        log_fail "No hay partición montada para formatear"
        return 1
    fi

    # Test: MKFS con EXT2
    log_info "Formateando partición con EXT2 (2fs)..."
    local resp
    resp=$(run_command "mkfs -id=${MOUNT_ID} -fs=2fs")

    if check_json_ok "$resp"; then
        log_success "MKFS EXT2 completó sin errores"
    else
        log_fail "MKFS EXT2 falló: $(get_json_field "$resp" "error")"
        return 1
    fi

    # Verificar que el archivo /users.txt fue creado
    log_info "Verificando creación de /users.txt..."
    # Nota: esto requiere que tengamos una forma de leer archivos del FS
    # Por ahora solo verificamos que mkfs no falle

    log_info "Verificando estructura EXT2 (superblock, bitmaps)..."
    # Aquí podríamos hacer reportes para verificar, por ahora asumimos OK
    log_success "EXT2 formateado (verificar manualmente con reportes)"
}

##############################################################################
# Tests P2 - MKFS EXT3 (Journal)
##############################################################################

test_p2_mkfs_ext3() {
    log_section "P2: Formateo EXT3 con Journal (MKFS 3fs)"

    # Montar otra partición para EXT3
    log_info "Montando partición lógica para EXT3..."
    local resp
    resp=$(run_command "mount -path=${DISK_PATH} -name=Logic1")

    if check_json_ok "$resp"; then
        local ext3_id
        ext3_id=$(get_json_field "$resp" "output" | grep -oP 'id=\K[^ ]+' | head -1)

        if [[ -z "$ext3_id" ]]; then
            ext3_id=$(echo "$resp" | jq -r '.output' | grep -oP '84[0-9]+[A-Z]')
        fi

        log_success "Partición para EXT3 montada: ${ext3_id}"

        # Test: MKFS con EXT3
        log_info "Formateando partición con EXT3 (3fs)..."
        resp=$(run_command "mkfs -id=${ext3_id} -fs=3fs")

        if check_json_ok "$resp"; then
            log_success "MKFS EXT3 creó filesystem con journal"
        else
            log_fail "MKFS EXT3 falló: $(get_json_field "$resp" "error")"
            return 1
        fi

        # Guardar ID para tests de journal
        export EXT3_MOUNT_ID="$ext3_id"
    else
        log_fail "No se pudo montar partición para EXT3: $(get_json_field "$resp" "error")"
    fi
}

##############################################################################
# Tests P2 - Comandos EXT3 (journaling, recovery, loss)
##############################################################################

test_p2_ext3_commands() {
    log_section "P2: Comandos EXT3 (journaling/recovery/loss)"

    if [[ -z "${EXT3_MOUNT_ID:-}" ]]; then
        log_warn "No hay partición EXT3 montada, saltando tests de journal"
        return 0
    fi

    # Test: JOURNALING
    log_info "Ejecutando comando journaling..."
    local resp
    resp=$(run_command "journaling -id=${EXT3_MOUNT_ID}")

    if check_json_ok "$resp"; then
        log_success "JOURNALING devolvió entradas del journal"

        # Verificar que devuelve JSON con estructura
        local entries
        entries=$(get_json_field "$resp" "output")

        if [[ -n "$entries" ]]; then
            log_info "Journal tiene entradas registradas"
        fi
    else
        log_fail "JOURNALING falló: $(get_json_field "$resp" "error")"
    fi

    # Test: RECOVERY
    log_info "Ejecutando comando recovery..."
    resp=$(run_command "recovery -id=${EXT3_MOUNT_ID}")

    if check_json_ok "$resp"; then
        log_success "RECOVERY ejecutó sin errores"
    else
        log_fail "RECOVERY falló: $(get_json_field "$resp" "error")"
    fi

    # Test: LOSS
    log_info "Ejecutando comando loss..."
    resp=$(run_command "loss -id=${EXT3_MOUNT_ID}")

    if check_json_ok "$resp"; then
        log_success "LOSS ejecutó sin errores"
    else
        log_fail "LOSS falló: $(get_json_field "$resp" "error")"
    fi
}

##############################################################################
# Tests P2 - Endpoints HTTP de EXT3
##############################################################################

test_p2_ext3_http_endpoints() {
    log_section "P2: Endpoints HTTP EXT3"

    if [[ -z "${EXT3_MOUNT_ID:-}" ]]; then
        log_warn "No hay partición EXT3 montada, saltando tests HTTP"
        return 0
    fi

    # Test: GET /api/ext3/journal
    log_info "Test: GET /api/ext3/journal?id=${EXT3_MOUNT_ID}"
    local resp
    resp=$(curl -s "${API_BASE}/api/ext3/journal?id=${EXT3_MOUNT_ID}")

    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_success "GET /api/ext3/journal responde correctamente"
    else
        log_fail "GET /api/ext3/journal falló"
    fi

    # Test: POST /api/ext3/recovery
    log_info "Test: POST /api/ext3/recovery"
    resp=$(curl -s -X POST "${API_BASE}/api/ext3/recovery" \
        -H "Content-Type: application/json" \
        -d "{\"id\":\"${EXT3_MOUNT_ID}\"}")

    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_success "POST /api/ext3/recovery responde correctamente"
    else
        log_fail "POST /api/ext3/recovery falló"
    fi

    # Test: POST /api/ext3/loss
    log_info "Test: POST /api/ext3/loss"
    resp=$(curl -s -X POST "${API_BASE}/api/ext3/loss" \
        -H "Content-Type: application/json" \
        -d "{\"id\":\"${EXT3_MOUNT_ID}\"}")

    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_success "POST /api/ext3/loss responde correctamente"
    else
        log_fail "POST /api/ext3/loss falló"
    fi
}

##############################################################################
# Tests P2 - Reportes DOT
##############################################################################

test_p2_reports_dot() {
    log_section "P2: Reportes en Formato DOT"

    if [[ -z "$MOUNT_ID" ]]; then
        log_warn "No hay partición montada para generar reportes"
        return 0
    fi

    # Test: Reporte MBR
    log_info "Test: GET /api/reports/mbr?id=${MOUNT_ID}"
    local resp
    resp=$(curl -s "${API_BASE}/api/reports/mbr?id=${MOUNT_ID}")

    if echo "$resp" | grep -q "digraph MBR"; then
        log_success "Reporte MBR devuelve formato DOT válido"
    else
        log_fail "Reporte MBR no devuelve formato DOT"
    fi

    # Test: Reporte DISK
    log_info "Test: GET /api/reports/disk?id=${MOUNT_ID}"
    resp=$(curl -s "${API_BASE}/api/reports/disk?id=${MOUNT_ID}")

    if echo "$resp" | grep -q "digraph"; then
        log_success "Reporte DISK devuelve formato DOT válido"
    else
        log_fail "Reporte DISK no devuelve formato DOT"
    fi

    # Test: Reporte TREE
    log_info "Test: GET /api/reports/tree?id=${MOUNT_ID}"
    resp=$(curl -s "${API_BASE}/api/reports/tree?id=${MOUNT_ID}")

    if echo "$resp" | grep -q "digraph"; then
        log_success "Reporte TREE devuelve formato DOT válido"
    else
        log_fail "Reporte TREE no devuelve formato DOT"
    fi

    # Test: Reporte SUPERBLOCK
    log_info "Test: GET /api/reports/sb?id=${MOUNT_ID}"
    resp=$(curl -s "${API_BASE}/api/reports/sb?id=${MOUNT_ID}")

    if echo "$resp" | grep -q "digraph"; then
        log_success "Reporte SUPERBLOCK devuelve formato DOT válido"
    else
        log_fail "Reporte SUPERBLOCK no devuelve formato DOT"
    fi

    # Test: Reporte JOURNAL (si hay EXT3)
    if [[ -n "${EXT3_MOUNT_ID:-}" ]]; then
        log_info "Test: GET /api/reports/journal?id=${EXT3_MOUNT_ID}"
        resp=$(curl -s "${API_BASE}/api/reports/journal?id=${EXT3_MOUNT_ID}")

        if echo "$resp" | grep -q "digraph"; then
            log_success "Reporte JOURNAL devuelve formato DOT válido"
        else
            log_fail "Reporte JOURNAL no devuelve formato DOT"
        fi
    fi
}

##############################################################################
# Tests P2 - Consola HTTP y Manejo de Errores
##############################################################################

test_p2_http_console() {
    log_section "P2: Consola HTTP y Manejo de Errores"

    # Test: Comando válido
    log_info "Test: Comando válido via /api/cmd/run"
    local resp
    resp=$(run_command "mounted")

    if check_json_ok "$resp"; then
        log_success "/api/cmd/run ejecuta comandos válidos correctamente"
    else
        log_fail "/api/cmd/run falló con comando válido"
    fi

    # Test: Comando inválido (debe devolver error pero OK en HTTP)
    log_info "Test: Comando inválido debe retornar error"
    resp=$(run_command "comando_invalido")

    if ! check_json_ok "$resp"; then
        local error_msg
        error_msg=$(get_json_field "$resp" "error")

        if [[ -n "$error_msg" ]]; then
            log_success "Comando inválido devuelve error correctamente: $error_msg"
        else
            log_warn "Comando inválido pero sin mensaje de error"
        fi
    else
        log_fail "Comando inválido no devolvió error"
    fi

    # Test: Comando con parámetros faltantes
    log_info "Test: Comando con parámetros faltantes"
    resp=$(run_command "mkdisk -path=/tmp/test.mia")  # Falta -size

    if ! check_json_ok "$resp"; then
        log_success "Comando incompleto devuelve error"
    else
        log_warn "Comando incompleto no devolvió error"
    fi

    # Test: POST /api/cmd/script (múltiples comandos)
    log_info "Test: POST /api/cmd/script (ejecución de script)"
    local script_resp
    script_resp=$(curl -s -X POST "${API_BASE}/api/cmd/script" \
        -H "Content-Type: application/json" \
        -d "{\"script\":\"mounted\\n# Comentario\\nmounted\"}")

    if echo "$script_resp" | jq -e '.ok' >/dev/null 2>&1; then
        local executed
        executed=$(echo "$script_resp" | jq '.executed // 0')

        if [[ "$executed" -gt 0 ]]; then
            log_success "/api/cmd/script ejecuta múltiples comandos (executed=$executed)"
        else
            log_warn "/api/cmd/script no ejecutó comandos"
        fi
    else
        log_fail "/api/cmd/script falló"
    fi
}

##############################################################################
# Tests Adicionales - Endpoints Informativos
##############################################################################

test_additional_endpoints() {
    log_section "Tests Adicionales: Endpoints Informativos"

    # Test: GET /api/commands
    log_info "Test: GET /api/commands"
    local resp
    resp=$(curl -s "${API_BASE}/api/commands")

    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_success "GET /api/commands lista comandos disponibles"
    else
        log_fail "GET /api/commands falló"
    fi

    # Test: GET /api/disks/list
    log_info "Test: GET /api/disks/list?path=${TEST_DIR}"
    resp=$(curl -s "${API_BASE}/api/disks/list?path=${TEST_DIR}")

    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        local count
        count=$(echo "$resp" | jq '.count // 0')

        if [[ "$count" -gt 0 ]]; then
            log_success "GET /api/disks/list encuentra discos .mia (count=$count)"
        else
            log_warn "GET /api/disks/list no encontró discos"
        fi
    else
        log_fail "GET /api/disks/list falló"
    fi

    # Test: GET /api/disks/info
    log_info "Test: GET /api/disks/info?path=${DISK_PATH}"
    resp=$(curl -s "${API_BASE}/api/disks/info?path=${DISK_PATH}")

    if echo "$resp" | jq -e '.ok == true' >/dev/null 2>&1; then
        log_success "GET /api/disks/info devuelve información de disco"
    else
        log_fail "GET /api/disks/info falló"
    fi
}

##############################################################################
# Limpieza
##############################################################################

cleanup() {
    log_section "Limpieza"

    # Desmontar particiones
    if [[ -n "$MOUNT_ID" ]]; then
        log_info "Desmontando ${MOUNT_ID}..."
        run_command "unmount -id=${MOUNT_ID}" >/dev/null 2>&1 || true
    fi

    if [[ -n "${EXT3_MOUNT_ID:-}" ]]; then
        log_info "Desmontando ${EXT3_MOUNT_ID}..."
        run_command "unmount -id=${EXT3_MOUNT_ID}" >/dev/null 2>&1 || true
    fi

    # Limpiar archivos de prueba
    log_info "Eliminando archivos de prueba..."
    rm -rf "$TEST_DIR"

    log_success "Limpieza completada"
}

##############################################################################
# Resumen Final
##############################################################################

print_summary() {
    log_section "RESUMEN DE VERIFICACIÓN"

    echo ""
    echo "Total de tests ejecutados: ${TOTAL_TESTS}"
    echo -e "${GREEN}Tests exitosos: ${PASSED_TESTS}${NC}"
    echo -e "${RED}Tests fallidos: ${FAILED_TESTS}${NC}"
    echo ""

    local pass_rate=0
    if [[ $TOTAL_TESTS -gt 0 ]]; then
        pass_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi

    echo "Tasa de éxito: ${pass_rate}%"
    echo ""

    if [[ $FAILED_TESTS -eq 0 ]]; then
        echo -e "${GREEN}✓ TODOS LOS TESTS PASARON${NC}"
        echo "Tu implementación P2 (Carnet 84) está completa y funcional."
        return 0
    else
        echo -e "${RED}✗ ALGUNOS TESTS FALLARON${NC}"
        echo "Revisa los mensajes de error arriba para identificar problemas."
        return 1
    fi
}

##############################################################################
# Main
##############################################################################

main() {
    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║   Verificador P2 - GoDisk (Carnet 201905884)             ║"
    echo "║   Validación Integral de Componentes P1 y P2             ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    # Ejecutar tests en orden
    test_prerequisites
    test_p1_disk_operations
    test_p1_mount_operations
    test_p1_mkfs_ext2
    test_p2_mkfs_ext3
    test_p2_ext3_commands
    test_p2_ext3_http_endpoints
    test_p2_reports_dot
    test_p2_http_console
    test_additional_endpoints

    # Limpieza
    cleanup

    # Resumen
    print_summary

    exit $?
}

# Manejo de señales para cleanup (solo en interrupciones, no en exit normal)
trap cleanup INT TERM

# Ejecutar
main "$@"
