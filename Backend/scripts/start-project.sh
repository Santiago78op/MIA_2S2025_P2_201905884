#!/bin/bash

# GoDisk Project Startup Script
# Este script inicia tanto el backend como el frontend en paralelo

echo "🚀 Iniciando GoDisk Project..."
echo "================================"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Obtener el directorio del proyecto (dos niveles arriba del script)
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd ../.. && pwd)"
BACKEND_DIR="$PROJECT_DIR/Backend"
FRONTEND_DIR="$PROJECT_DIR/Frontend/godisk-frontend"

echo -e "${BLUE}📁 Directorio del proyecto: $PROJECT_DIR${NC}"

# Función para verificar si un puerto está en uso
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        return 0  # Puerto en uso
    else
        return 1  # Puerto libre
    fi
}

# Función para matar procesos en puertos específicos
kill_port_processes() {
    local port=$1
    local process_name=$2
    
    if check_port $port; then
        echo -e "${YELLOW}⚠️  Puerto $port está en uso. Matando procesos de $process_name...${NC}"
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
        sleep 2
    fi
}

# Verificar y limpiar puertos antes de iniciar
echo -e "${YELLOW}🔍 Verificando puertos...${NC}"
kill_port_processes 8080 "Backend (Go)"
kill_port_processes 5173 "Frontend (Vite)"
kill_port_processes 3000 "Frontend (alternativo)"

# Verificar que existan los directorios
if [ ! -d "$BACKEND_DIR" ]; then
    echo -e "${RED}❌ Error: No se encontró el directorio del backend: $BACKEND_DIR${NC}"
    exit 1
fi

if [ ! -d "$FRONTEND_DIR" ]; then
    echo -e "${RED}❌ Error: No se encontró el directorio del frontend: $FRONTEND_DIR${NC}"
    exit 1
fi

# Verificar dependencias
echo -e "${BLUE}🔍 Verificando dependencias...${NC}"

# Verificar Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go no está instalado. Por favor instala Go primero.${NC}"
    exit 1
fi

# Verificar Node.js
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js no está instalado. Por favor instala Node.js primero.${NC}"
    exit 1
fi

# Verificar npm
if ! command -v npm &> /dev/null; then
    echo -e "${RED}❌ npm no está instalado. Por favor instala npm primero.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Todas las dependencias están instaladas${NC}"

# Función para iniciar el backend
start_backend() {
    echo -e "${BLUE}🔧 Iniciando Backend (Go)...${NC}"
    cd "$BACKEND_DIR/cmd/server"
    
    # Verificar que go.mod existe
    if [ ! -f "$BACKEND_DIR/go.mod" ]; then
        echo -e "${RED}❌ Error: go.mod no encontrado en $BACKEND_DIR${NC}"
        return 1
    fi
    
    # Descargar dependencias si es necesario
    echo -e "${YELLOW}📦 Descargando dependencias de Go...${NC}"
    cd "$BACKEND_DIR"
    go mod download
    
    # Iniciar el servidor
    cd "$BACKEND_DIR/cmd/server"
    echo -e "${GREEN}🟢 Backend iniciando en puerto 8080...${NC}"
    go run . 2>&1 | sed "s/^/[BACKEND] /"
}

# Función para iniciar el frontend
start_frontend() {
    echo -e "${BLUE}🔧 Iniciando Frontend (React + Vite)...${NC}"
    cd "$FRONTEND_DIR"
    
    # Verificar que package.json existe
    if [ ! -f "package.json" ]; then
        echo -e "${RED}❌ Error: package.json no encontrado en $FRONTEND_DIR${NC}"
        return 1
    fi
    
    # Instalar dependencias si node_modules no existe
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}📦 Instalando dependencias de npm...${NC}"
        npm install
    fi
    
    # Iniciar el servidor de desarrollo
    echo -e "${GREEN}🟢 Frontend iniciando en puerto 5173...${NC}"
    npm run dev 2>&1 | sed "s/^/[FRONTEND] /"
}

# Función para manejar la señal de interrupción (Ctrl+C)
cleanup() {
    echo -e "\n${YELLOW}🛑 Deteniendo servicios...${NC}"
    
    # Matar procesos del backend y frontend
    pkill -f "go run"
    pkill -f "vite"
    pkill -f "node.*vite"
    
    # Matar procesos en los puertos específicos
    kill_port_processes 8080 "Backend"
    kill_port_processes 5173 "Frontend"
    
    echo -e "${GREEN}✅ Servicios detenidos correctamente${NC}"
    exit 0
}

# Configurar trap para manejar Ctrl+C
trap cleanup SIGINT SIGTERM

# Crear directorios de logs si no existen
mkdir -p "$PROJECT_DIR/logs"

# Iniciar servicios en paralelo
echo -e "${GREEN}🚀 Iniciando servicios...${NC}"
echo -e "${BLUE}💡 Presiona Ctrl+C para detener todos los servicios${NC}"
echo "================================"

# Iniciar backend en background
start_backend > "$PROJECT_DIR/logs/backend.log" 2>&1 &
BACKEND_PID=$!

# Esperar un momento para que el backend inicie
sleep 3

# Iniciar frontend en background
start_frontend > "$PROJECT_DIR/logs/frontend.log" 2>&1 &
FRONTEND_PID=$!

# Mostrar información de los servicios
echo -e "${GREEN}✅ Servicios iniciados:${NC}"
echo -e "${BLUE}   📡 Backend (Go): http://localhost:8080${NC}"
echo -e "${BLUE}   🌐 Frontend (React): http://localhost:5173${NC}"
echo -e "${YELLOW}   📝 Logs: $PROJECT_DIR/logs/${NC}"
echo ""
echo -e "${BLUE}💡 Moniteando servicios... (Ctrl+C para detener)${NC}"

# Función para mostrar logs en tiempo real
show_logs() {
    # Mostrar logs intercalados
    tail -f "$PROJECT_DIR/logs/backend.log" "$PROJECT_DIR/logs/frontend.log" 2>/dev/null | while read line; do
        echo "$line"
    done
}

# Mostrar logs en tiempo real
show_logs &
LOGS_PID=$!

# Esperar a que los procesos terminen o sean interrumpidos
wait $BACKEND_PID $FRONTEND_PID

# Limpiar al terminar
kill $LOGS_PID 2>/dev/null || true
cleanup