#!/usr/bin/env bash
# Smoke Test P1 - Basado en el script del calificador
# Carnet: 201905884

set -euo pipefail

API=${API:-http://localhost:8080}
BASE_PATH="/home/julian/Documents/MIA_2S2025_P1_201905884/Discos"

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0

echo "=========================================="
echo "🧪 Smoke Test P1 - Carnet 84"
echo "=========================================="
echo ""

# Función para ejecutar comando
post() {
    local cmd="$1"
    local desc="${2:-$cmd}"

    echo -n "Testing: $desc... "

    response=$(curl -sS -X POST "$API/api/cmd/run" \
         -H 'Content-Type: application/json' \
         -d "{\"line\":\"$cmd\"}" 2>&1)

    if echo "$response" | jq -e '.ok == true' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "  Response: $(echo "$response" | jq -r '.error // .output // .' 2>/dev/null | head -c 200)"
        ((FAIL++))
        return 1
    fi
}

# Función para verificar ID de montaje
check_mount_id() {
    local expected_id="$1"
    local desc="$2"

    echo -n "Verifying: $desc has ID $expected_id... "

    response=$(curl -sS -X POST "$API/api/cmd/run" \
         -H 'Content-Type: application/json' \
         -d '{"line":"mounted"}')

    if echo "$response" | jq -r '.output' | grep -q "$expected_id"; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "  Expected ID: $expected_id"
        echo "  Output: $(echo "$response" | jq -r '.output' | head -c 200)"
        ((FAIL++))
        return 1
    fi
}

# Limpiar discos anteriores
echo "📋 Limpieza de discos anteriores..."
rm -rf "$BASE_PATH"
mkdir -p "$BASE_PATH"
echo ""

# ========== MKDISK ==========
echo "📋 Fase 1: MKDISK"
echo "----------------------------------------"
post "mkdisk -size=50 -unit=M -fit=FF -path=\"$BASE_PATH/Disco1.mia\"" "Create Disco1 (50MB)"
post "mkdisk -size=13 -path=\"$BASE_PATH/Disco3.mia\"" "Create Disco3 (13MB)"
post "mkdisk -size=20 -unit=M -fit=WF -path=\"$BASE_PATH/Disco5.mia\"" "Create Disco5 (20MB)"
echo ""

# ========== FDISK Disco1 ==========
echo "📋 Fase 2: FDISK Disco1 (4 primarias)"
echo "----------------------------------------"
post "fdisk -type=P -unit=b -name=Part11 -size=10485760 -path=\"$BASE_PATH/Disco1.mia\" -fit=BF" "Disco1 Part11 (10MB bytes)"
post "fdisk -type=P -unit=k -name=Part12 -size=10240 -path=\"$BASE_PATH/Disco1.mia\" -fit=BF" "Disco1 Part12 (10MB KB)"
post "fdisk -type=P -unit=M -name=Part13 -size=10 -path=\"$BASE_PATH/Disco1.mia\" -fit=BF" "Disco1 Part13 (10MB)"
post "fdisk -type=P -unit=b -name=Part14 -size=10485760 -path=\"$BASE_PATH/Disco1.mia\" -fit=BF" "Disco1 Part14 (10MB bytes)"
echo ""

# ========== FDISK Disco3 ==========
echo "📋 Fase 3: FDISK Disco3 (3 primarias)"
echo "----------------------------------------"
post "fdisk -type=P -unit=m -name=Part31 -size=4 -path=\"$BASE_PATH/Disco3.mia\"" "Disco3 Part31 (4MB)"
post "fdisk -type=P -unit=m -name=Part32 -size=4 -path=\"$BASE_PATH/Disco3.mia\"" "Disco3 Part32 (4MB)"
post "fdisk -type=P -unit=m -name=Part33 -size=1 -path=\"$BASE_PATH/Disco3.mia\"" "Disco3 Part33 (1MB)"
echo ""

# ========== FDISK Disco5 ==========
echo "📋 Fase 4: FDISK Disco5 (Extendida + Lógicas + Primaria)"
echo "----------------------------------------"
post "fdisk -type=E -unit=k -name=Part51 -size=5120 -path=\"$BASE_PATH/Disco5.mia\" -fit=BF" "Disco5 Part51 Extended (5MB)"
post "fdisk -type=L -unit=k -name=Part52 -size=1024 -path=\"$BASE_PATH/Disco5.mia\" -fit=BF" "Disco5 Part52 Logical (1MB)"
post "fdisk -type=P -unit=k -name=Part53 -size=5120 -path=\"$BASE_PATH/Disco5.mia\" -fit=BF" "Disco5 Part53 Primary (5MB)"
post "fdisk -type=L -unit=k -name=Part54 -size=1024 -path=\"$BASE_PATH/Disco5.mia\" -fit=BF" "Disco5 Part54 Logical (1MB)"
post "fdisk -type=L -unit=k -name=Part55 -size=1024 -path=\"$BASE_PATH/Disco5.mia\" -fit=BF" "Disco5 Part55 Logical (1MB)"
post "fdisk -type=L -unit=k -name=Part56 -size=1024 -path=\"$BASE_PATH/Disco5.mia\" -fit=BF" "Disco5 Part56 Logical (1MB)"
echo ""

# ========== MOUNT (IDs con terminación 84) ==========
echo "📋 Fase 5: MOUNT - Validación de IDs con terminación 84"
echo "----------------------------------------"
post "mount -path=\"$BASE_PATH/Disco1.mia\" -name=Part11" "Mount Disco1 Part11"
check_mount_id "841A" "Disco1 Part11"

post "mount -path=\"$BASE_PATH/Disco1.mia\" -name=Part12" "Mount Disco1 Part12"
check_mount_id "842A" "Disco1 Part12"

post "mount -path=\"$BASE_PATH/Disco3.mia\" -name=Part31" "Mount Disco3 Part31"
check_mount_id "841B" "Disco3 Part31"

post "mount -path=\"$BASE_PATH/Disco3.mia\" -name=Part32" "Mount Disco3 Part32"
check_mount_id "842B" "Disco3 Part32"

post "mount -path=\"$BASE_PATH/Disco5.mia\" -name=Part53" "Mount Disco5 Part53"
check_mount_id "841C" "Disco5 Part53"

post "mounted" "List all mounted partitions"
echo ""

# ========== MKFS ==========
echo "📋 Fase 6: MKFS EXT2 (debe crear /users.txt)"
echo "----------------------------------------"
post "mkfs -type=full -id=841A" "Format 841A with EXT2"
echo ""

# ========== LOGIN/ADMIN ==========
echo "📋 Fase 7: Login y Administración de Usuarios"
echo "----------------------------------------"
post "login -user=root -pass=123 -id=841A" "Login as root"
post "mkgrp -name=usuarios" "Create group 'usuarios'"
post "mkusr -user=user1 -pass=pass1 -grp=usuarios" "Create user 'user1'"
post "cat -file1=/users.txt" "Read /users.txt"
post "logout" "Logout"
echo ""

# ========== UNMOUNT ==========
echo "📋 Fase 8: UNMOUNT"
echo "----------------------------------------"
post "unmount -id=841A" "Unmount 841A"
post "unmount -id=842A" "Unmount 842A"
post "unmount -id=841B" "Unmount 841B"
post "unmount -id=842B" "Unmount 842B"
post "unmount -id=841C" "Unmount 841C"
echo ""

# ========== RESULTADOS ==========
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
