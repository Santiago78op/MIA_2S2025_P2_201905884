#!/bin/bash

# Script de conveniencia para matar procesos de GoDisk desde la raíz del proyecto
# Simplemente llama al script principal en Backend/scripts/

echo "🔴 GoDisk Process Killer"
echo "======================="

# Obtener el directorio del script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAIN_SCRIPT="$SCRIPT_DIR/Backend/scripts/kill-processes.sh"

# Verificar que el script principal existe
if [ ! -f "$MAIN_SCRIPT" ]; then
    echo "❌ Error: No se encontró el script principal en $MAIN_SCRIPT"
    exit 1
fi

# Ejecutar el script principal
echo "🔄 Ejecutando script de limpieza..."
exec "$MAIN_SCRIPT" "$@"