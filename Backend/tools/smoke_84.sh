#!/bin/bash
# GoDisk 2.0 - Smoke Test Completo (Carnet: 201905884)
# Autor: Santiago Julian Barrera Reyes
# Este script valida todas las funcionalidades críticas del sistema

set -e  # Salir en caso de error

DISK_PATH="/tmp/Disk84.mia"
API_URL="http://localhost:8080"
PASS=0
FAIL=0

# Colores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "🧪 GoDisk 2.0 - Smoke Test (vd84)"
echo "=========================================="
echo ""

# Función auxiliar para ejecutar comandos
run_cmd() {
    local cmd="$1"
    local desc="$2"

    echo -n "Testing: $desc... "

    response=$(curl -s -X POST "$API_URL/api/cmd/run" \
        -H "Content-Type: application/json" \
        -d "{\"line\": \"$cmd\"}")

    if echo "$response" | jq -e '.ok == true' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "  Response: $response" | head -c 200
        echo ""
        ((FAIL++))
        return 1
    fi
}

# Función para verificar reportes DOT
check_dot_report() {
    local endpoint="$1"
    local desc="$2"

    echo -n "Testing: $desc... "

    response=$(curl -s "$API_URL/api/reports/$endpoint")

    if echo "$response" | grep -q "digraph"; then
        echo -e "${GREEN}✓ PASS (DOT format)${NC}"
        ((PASS++))
        return 0
    else
        echo -e "${RED}✗ FAIL (not DOT)${NC}"
        ((FAIL++))
        return 1
    fi
}

# Limpiar disco anterior si existe
rm -f "$DISK_PATH"

echo "📋 Fase 1: Gestión de Discos y Particiones"
echo "----------------------------------------"

# 1. Crear disco
run_cmd "mkdisk -path=\"$DISK_PATH\" -size=20 -unit=m -fit=ff" \
        "Create 20MB disk"

# 2. Crear partición primaria
run_cmd "fdisk -mode=add -path=\"$DISK_PATH\" -name=PRIM1 -size=1024 -unit=k -type=p -fit=ff" \
        "Add primary partition PRIM1"

# 3. Crear partición extendida
run_cmd "fdisk -mode=add -path=\"$DISK_PATH\" -name=EXT1 -size=8 -unit=m -type=e -fit=ff" \
        "Add extended partition EXT1"

# 4. Crear partición lógica
run_cmd "fdisk -mode=add -path=\"$DISK_PATH\" -name=LOG1 -size=1024 -unit=k -type=l -fit=ff" \
        "Add logical partition LOG1"

echo ""
echo "📋 Fase 2: Montajes y Sistema de IDs"
echo "----------------------------------------"

# 5. Montar PRIM1 (debe generar vd84)
run_cmd "mount -path=\"$DISK_PATH\" -name=PRIM1" \
        "Mount PRIM1 (should get vd84)"

# 6. Montar LOG1 (debe generar vd841)
run_cmd "mount -path=\"$DISK_PATH\" -name=LOG1" \
        "Mount LOG1 (should get vd841)"

# 7. Verificar montajes
echo -n "Testing: List mounted partitions (mounted command)... "
mounted_out=$(curl -s -X POST "$API_URL/api/cmd/run" \
    -H "Content-Type: application/json" \
    -d '{"line": "mounted"}')

if echo "$mounted_out" | grep -q "vd84"; then
    echo -e "${GREEN}✓ PASS (vd84 found)${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL (vd84 not found)${NC}"
    ((FAIL++))
fi

echo ""
echo "📋 Fase 3: Formateo EXT2"
echo "----------------------------------------"

# 8. Formatear con EXT2
run_cmd "mkfs -id=vd84 -fs=2fs" \
        "Format vd84 with EXT2"

echo ""
echo "📋 Fase 4: Formateo EXT3"
echo "----------------------------------------"

# 9. Formatear con EXT3
run_cmd "mkfs -id=vd841 -fs=3fs" \
        "Format vd841 with EXT3"

echo ""
echo "📋 Fase 5: Reportes DOT"
echo "----------------------------------------"

# 10. Reporte MBR
check_dot_report "mbr?path=$DISK_PATH" \
                  "MBR report returns DOT format"

# 11. Reporte Tree
check_dot_report "tree?id=vd84" \
                  "Tree report returns DOT format"

# 12. Reporte Journal
check_dot_report "journal?id=vd841" \
                  "Journal report returns DOT format"

echo ""
echo "📋 Fase 6: Operaciones de Archivos (EXT3)"
echo "----------------------------------------"

# 13. Crear directorio
run_cmd "mkdir -id=vd841 -path=\"/docs\" -p" \
        "Create directory /docs"

# 14. Crear archivo
run_cmd "mkfile -id=vd841 -path=\"/docs/test.txt\" -cont=\"Hello GoDisk 2.0\"" \
        "Create file /docs/test.txt"

# 15. Editar archivo
run_cmd "edit -id=vd841 -path=\"/docs/test.txt\" -cont=\" - Edited\"" \
        "Edit file (append mode)"

echo ""
echo "📋 Fase 7: Journal EXT3"
echo "----------------------------------------"

# 16. Verificar journal
echo -n "Testing: Journal contains operations... "
journal_out=$(curl -s -X POST "$API_URL/api/cmd/run" \
    -H "Content-Type: application/json" \
    -d '{"line": "journaling -id=vd841"}')

if echo "$journal_out" | grep -q "mkfs"; then
    echo -e "${GREEN}✓ PASS (journal has entries)${NC}"
    ((PASS++))
else
    echo -e "${RED}✗ FAIL (journal empty)${NC}"
    ((FAIL++))
fi

echo ""
echo "📋 Fase 8: Desmontaje"
echo "----------------------------------------"

# 17. Desmontar particiones
run_cmd "unmount -id=vd84" \
        "Unmount vd84"

run_cmd "unmount -id=vd841" \
        "Unmount vd841"

echo ""
echo "=========================================="
echo "📊 RESULTADOS FINALES"
echo "=========================================="
echo -e "Pruebas exitosas: ${GREEN}$PASS${NC}"
echo -e "Pruebas fallidas:  ${RED}$FAIL${NC}"
echo "Total: $((PASS + FAIL))"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ TODOS LOS TESTS PASARON${NC}"
    exit 0
else
    echo -e "${RED}❌ ALGUNOS TESTS FALLARON${NC}"
    exit 1
fi
