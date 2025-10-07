#!/bin/bash

# Script de conveniencia para iniciar GoDisk desde la raíz del proyecto
# Simplemente llama al script principal en Backend/scripts/

echo "🚀 GoDisk Launcher"
echo "=================="

# Obtener el directorio del script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAIN_SCRIPT="$SCRIPT_DIR/Backend/scripts/start-project.sh"

# Verificar que el script principal existe
if [ ! -f "$MAIN_SCRIPT" ]; then
    echo "❌ Error: No se encontró el script principal en $MAIN_SCRIPT"
    exit 1
fi

# Ejecutar el script principal
echo "🔄 Ejecutando script principal..."
exec "$MAIN_SCRIPT" "$@"