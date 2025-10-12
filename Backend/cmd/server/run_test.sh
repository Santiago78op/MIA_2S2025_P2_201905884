#!/bin/bash

# Script auxiliar para ejecutar scripts de prueba SMIA
# Uso: ./run_test.sh <archivo.smia>

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para imprimir mensajes
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Verificar argumentos
if [ -z "$1" ]; then
    print_error "Falta el archivo de script"
    echo "Uso: $0 <script.smia>"
    echo ""
    echo "Ejemplo:"
    echo "  $0 test_basico.smia"
    exit 1
fi

SCRIPT_FILE=$1

# Verificar que el archivo existe
if [ ! -f "$SCRIPT_FILE" ]; then
    print_error "Archivo no encontrado: $SCRIPT_FILE"
    exit 1
fi

# Verificar que jq está instalado
if ! command -v jq &> /dev/null; then
    print_error "jq no está instalado"
    echo "Instala jq con: sudo apt-get install jq"
    exit 1
fi

# Verificar que el servidor está corriendo
print_info "Verificando servidor..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    print_error "El servidor no está corriendo en el puerto 8080"
    echo ""
    echo "Inicia el servidor con:"
    echo "  cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend"
    echo "  ./server &"
    exit 1
fi

print_info "Servidor está corriendo ✓"

# Crear JSON temporal
TEMP_JSON=$(mktemp)
print_info "Creando payload JSON..."
jq -n --rawfile script "$SCRIPT_FILE" '{script: $script}' > "$TEMP_JSON"

# Ejecutar script
print_info "Ejecutando script: $SCRIPT_FILE"
echo ""
echo "═══════════════════════════════════════════════════════════════"
echo ""

RESPONSE=$(curl -s -X POST http://localhost:8080/api/script \
  -H "Content-Type: application/json" \
  -d @"$TEMP_JSON")

# Limpiar archivo temporal
rm "$TEMP_JSON"

# Mostrar respuesta formateada
echo "$RESPONSE" | jq -r '
  if .output then
    .output
  else
    . | tojson
  end
'

echo ""
echo "═══════════════════════════════════════════════════════════════"
echo ""

# Contar errores en la respuesta
ERROR_COUNT=$(echo "$RESPONSE" | jq -r '.output' | grep -c "ERROR" || true)
if [ "$ERROR_COUNT" -gt 0 ]; then
    print_warning "Se encontraron $ERROR_COUNT errores en la ejecución"
else
    print_info "Script ejecutado sin errores ✓"
fi

# Mostrar resumen de archivos generados
print_info "Archivos generados:"
echo ""
echo "Discos:"
ls -lh Discos/ 2>/dev/null | tail -n +2 || echo "  (ninguno)"
echo ""
echo "Reportes:"
ls -lh Reports/ 2>/dev/null | tail -n +2 || echo "  (ninguno)"
echo ""

print_info "Test completado"
