#!/bin/bash

# Script para ejecutar el test completo y generar reporte de inodos

# Asegurarse de que el directorio de reportes existe
mkdir -p /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/Reports

# Ejecutar el servidor en segundo plano con el script de prueba
cd /home/julian/Documents/MIA_2S2025_P2_201905884/Backend

# Compilar
echo "Compilando backend..."
go build -o backend_server

# Ejecutar
echo "Ejecutando backend con script de prueba..."
./backend_server < tests/test.smia

echo "Script ejecutado. Verificar reportes en /home/julian/Documents/MIA_2S2025_P2_201905884/Backend/Reports/"
