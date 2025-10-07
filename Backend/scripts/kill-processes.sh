#!/bin/bash

# GoDisk Process Killer Script
# Este script mata todos los procesos relacionados con el proyecto GoDisk

echo "🔴 GoDisk Process Killer"
echo "======================="

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para verificar si un puerto está en uso
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null ; then
        return 0  # Puerto en uso
    else
        return 1  # Puerto libre
    fi
}

# Función para matar procesos en un puerto específico
kill_port_processes() {
    local port=$1
    local service_name=$2
    
    echo -e "${BLUE}🔍 Verificando puerto $port ($service_name)...${NC}"
    
    if check_port $port; then
        echo -e "${YELLOW}⚠️  Encontrados procesos en puerto $port${NC}"
        
        # Mostrar qué procesos están usando el puerto
        echo -e "${BLUE}📋 Procesos encontrados:${NC}"
        lsof -Pi :$port -sTCP:LISTEN
        
        # Matar los procesos
        echo -e "${RED}💀 Matando procesos en puerto $port...${NC}"
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
        
        # Verificar que se mataron
        sleep 1
        if check_port $port; then
            echo -e "${RED}❌ Error: Algunos procesos en puerto $port siguen activos${NC}"
            # Intentar de nuevo con más fuerza
            lsof -ti:$port | xargs kill -KILL 2>/dev/null || true
            sleep 1
        fi
        
        if ! check_port $port; then
            echo -e "${GREEN}✅ Puerto $port liberado correctamente${NC}"
        else
            echo -e "${RED}❌ No se pudo liberar el puerto $port${NC}"
        fi
    else
        echo -e "${GREEN}✅ Puerto $port ya está libre${NC}"
    fi
    echo ""
}

# Función para matar procesos por nombre/comando
kill_process_by_name() {
    local process_pattern=$1
    local description=$2
    
    echo -e "${BLUE}🔍 Buscando procesos: $description${NC}"
    
    # Buscar procesos que coincidan con el patrón
    pids=$(pgrep -f "$process_pattern" 2>/dev/null || true)
    
    if [ -n "$pids" ]; then
        echo -e "${YELLOW}⚠️  Encontrados procesos: $description${NC}"
        
        # Mostrar información de los procesos
        echo -e "${BLUE}📋 Procesos encontrados:${NC}"
        ps -p $pids -o pid,ppid,cmd 2>/dev/null || true
        
        # Matar los procesos
        echo -e "${RED}💀 Matando procesos: $description${NC}"
        echo $pids | xargs kill -9 2>/dev/null || true
        
        # Verificar que se mataron
        sleep 1
        remaining_pids=$(pgrep -f "$process_pattern" 2>/dev/null || true)
        if [ -n "$remaining_pids" ]; then
            echo -e "${RED}❌ Algunos procesos siguen activos, usando SIGKILL...${NC}"
            echo $remaining_pids | xargs kill -KILL 2>/dev/null || true
            sleep 1
        fi
        
        # Verificación final
        final_pids=$(pgrep -f "$process_pattern" 2>/dev/null || true)
        if [ -z "$final_pids" ]; then
            echo -e "${GREEN}✅ Todos los procesos de '$description' fueron terminados${NC}"
        else
            echo -e "${RED}❌ Algunos procesos de '$description' siguen activos${NC}"
        fi
    else
        echo -e "${GREEN}✅ No se encontraron procesos: $description${NC}"
    fi
    echo ""
}

# Función principal
main() {
    echo -e "${BLUE}🔍 Iniciando búsqueda y eliminación de procesos GoDisk...${NC}"
    echo ""
    
    # 1. Matar procesos por puertos específicos
    echo -e "${YELLOW}📡 Verificando puertos del proyecto...${NC}"
    kill_port_processes 8080 "Backend (Go Server)"
    kill_port_processes 5173 "Frontend (Vite Dev Server)"
    kill_port_processes 3000 "Frontend (alternativo)"
    kill_port_processes 4173 "Frontend (Vite Preview)"
    
    # 2. Matar procesos específicos del backend
    echo -e "${YELLOW}🔧 Verificando procesos del Backend...${NC}"
    kill_process_by_name "go run.*server" "Go Server (go run)"
    kill_process_by_name "godisk-server" "GoDisk Server Binary"
    kill_process_by_name "server.*godisk" "GoDisk Server Process"
    
    # 3. Matar procesos específicos del frontend
    echo -e "${YELLOW}🌐 Verificando procesos del Frontend...${NC}"
    kill_process_by_name "vite.*godisk" "Vite Dev Server"
    kill_process_by_name "node.*vite.*godisk" "Node Vite Process"
    kill_process_by_name "npm.*dev.*godisk" "NPM Dev Process"
    
    # 4. Matar procesos de Node.js que puedan estar relacionados
    echo -e "${YELLOW}📦 Verificando procesos de Node.js relacionados...${NC}"
    kill_process_by_name "node.*godisk-frontend" "Node.js Frontend Process"
    
    # 5. Limpiar archivos de proceso si existen
    echo -e "${YELLOW}🧹 Limpiando archivos temporales...${NC}"
    PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd ../.. && pwd)"
    
    # Eliminar archivos .pid si existen
    find "$PROJECT_DIR" -name "*.pid" -type f -delete 2>/dev/null || true
    
    # Limpiar locks de npm si existen
    find "$PROJECT_DIR/Frontend/godisk-frontend" -name ".npmrc.lock" -type f -delete 2>/dev/null || true
    find "$PROJECT_DIR/Frontend/godisk-frontend" -name "package-lock.json.lock" -type f -delete 2>/dev/null || true
    
    echo -e "${GREEN}✅ Limpieza de archivos temporales completada${NC}"
    echo ""
    
    # 6. Verificación final
    echo -e "${BLUE}🔍 Verificación final...${NC}"
    
    # Verificar puertos
    ports_still_used=()
    for port in 8080 5173 3000 4173; do
        if check_port $port; then
            ports_still_used+=($port)
        fi
    done
    
    if [ ${#ports_still_used[@]} -eq 0 ]; then
        echo -e "${GREEN}✅ Todos los puertos están libres${NC}"
    else
        echo -e "${RED}❌ Los siguientes puertos siguen en uso: ${ports_still_used[*]}${NC}"
        echo -e "${YELLOW}💡 Puedes intentar reiniciar manualmente o usar 'sudo lsof -ti:PUERTO | xargs kill -9'${NC}"
    fi
    
    # Verificar procesos restantes
    remaining_go_processes=$(pgrep -f "go.*run.*server" 2>/dev/null || true)
    remaining_node_processes=$(pgrep -f "node.*vite.*godisk" 2>/dev/null || true)
    
    if [ -z "$remaining_go_processes" ] && [ -z "$remaining_node_processes" ]; then
        echo -e "${GREEN}✅ No se encontraron procesos restantes del proyecto${NC}"
    else
        echo -e "${YELLOW}⚠️  Algunos procesos pueden seguir activos${NC}"
        if [ -n "$remaining_go_processes" ]; then
            echo -e "${BLUE}   Go processes: $remaining_go_processes${NC}"
        fi
        if [ -n "$remaining_node_processes" ]; then
            echo -e "${BLUE}   Node processes: $remaining_node_processes${NC}"
        fi
    fi
    
    echo ""
    echo -e "${GREEN}🎉 Proceso de limpieza completado${NC}"
    echo -e "${BLUE}💡 Ahora puedes ejecutar start-project.sh sin conflictos${NC}"
}

# Verificar si se está ejecutando como root (opcional)
if [ "$EUID" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  Ejecutándose como root. Esto puede ser necesario para algunos procesos.${NC}"
    echo ""
fi

# Preguntar confirmación a menos que se pase el parámetro -f (force)
if [ "$1" != "-f" ] && [ "$1" != "--force" ]; then
    echo -e "${YELLOW}⚠️  Este script matará todos los procesos relacionados con GoDisk${NC}"
    echo -e "${BLUE}   - Backend Go Server (puerto 8080)${NC}"
    echo -e "${BLUE}   - Frontend Vite Server (puerto 5173)${NC}"
    echo -e "${BLUE}   - Cualquier proceso Node.js relacionado${NC}"
    echo ""
    read -p "¿Continuar? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${BLUE}❌ Operación cancelada${NC}"
        exit 0
    fi
    echo ""
fi

# Ejecutar función principal
main

# Opcional: mostrar puertos en uso después de la limpieza
echo ""
echo -e "${BLUE}📊 Estado actual de puertos relevantes:${NC}"
for port in 8080 5173 3000 4173; do
    if check_port $port; then
        echo -e "${RED}   Puerto $port: EN USO${NC}"
    else
        echo -e "${GREEN}   Puerto $port: LIBRE${NC}"
    fi
done