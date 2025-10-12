#!/bin/bash

# Script para gestionar el servidor MIA
# Uso: ./manage_server.sh {start|stop|restart|status|logs}

set -e

SERVER_DIR="/home/julian/Documents/MIA_2S2025_P2_201905884/Backend"
PID_FILE="$SERVER_DIR/server.pid"
LOG_FILE="$SERVER_DIR/server.log"
SERVER_BIN="$SERVER_DIR/server"

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Verificar si el servidor está corriendo
is_running() {
    if [ -f "$PID_FILE" ]; then
        PID=$(cat "$PID_FILE")
        if ps -p "$PID" > /dev/null 2>&1; then
            return 0
        fi
    fi
    return 1
}

# Iniciar servidor
start() {
    if is_running; then
        PID=$(cat "$PID_FILE")
        print_warning "El servidor ya está corriendo (PID: $PID)"
        return
    fi

    if [ ! -f "$SERVER_BIN" ]; then
        print_error "El binario del servidor no existe: $SERVER_BIN"
        echo "Compila el servidor con: go build -o server ./cmd/server"
        exit 1
    fi

    print_info "Iniciando servidor..."
    cd "$SERVER_DIR"
    ./server > "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"

    sleep 2

    if is_running; then
        PID=$(cat "$PID_FILE")
        print_info "Servidor iniciado correctamente (PID: $PID)"
        print_info "Escuchando en http://localhost:8080"
        print_info "Logs en: $LOG_FILE"
    else
        print_error "El servidor no pudo iniciarse"
        print_error "Revisa los logs en: $LOG_FILE"
        exit 1
    fi
}

# Detener servidor
stop() {
    if ! is_running; then
        print_warning "El servidor no está corriendo"
        if [ -f "$PID_FILE" ]; then
            rm "$PID_FILE"
        fi
        return
    fi

    PID=$(cat "$PID_FILE")
    print_info "Deteniendo servidor (PID: $PID)..."
    kill "$PID" 2>/dev/null || true

    # Esperar a que termine
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    # Si aún está corriendo, forzar
    if ps -p "$PID" > /dev/null 2>&1; then
        print_warning "Forzando detención..."
        kill -9 "$PID" 2>/dev/null || true
    fi

    rm "$PID_FILE"
    print_info "Servidor detenido"
}

# Reiniciar servidor
restart() {
    print_info "Reiniciando servidor..."
    stop
    sleep 2
    start
}

# Estado del servidor
status() {
    if is_running; then
        PID=$(cat "$PID_FILE")
        print_info "El servidor está corriendo (PID: $PID)"

        # Verificar conectividad
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_info "Servidor responde correctamente en http://localhost:8080"
        else
            print_warning "El servidor está corriendo pero no responde"
        fi

        # Mostrar uso de recursos
        echo ""
        echo "Uso de recursos:"
        ps -p "$PID" -o %cpu,%mem,rss,vsz,cmd | tail -n +2
    else
        print_warning "El servidor no está corriendo"
    fi
}

# Ver logs
logs() {
    if [ ! -f "$LOG_FILE" ]; then
        print_error "No se encontró archivo de logs: $LOG_FILE"
        exit 1
    fi

    if [ "$1" = "-f" ] || [ "$1" = "--follow" ]; then
        print_info "Mostrando logs (Ctrl+C para salir)..."
        tail -f "$LOG_FILE"
    else
        print_info "Últimas 50 líneas de logs:"
        tail -n 50 "$LOG_FILE"
    fi
}

# Compilar servidor
build() {
    print_info "Compilando servidor..."
    cd "$SERVER_DIR"
    go build -o server ./cmd/server
    print_info "Compilación exitosa"
    ls -lh server
}

# Mostrar ayuda
show_help() {
    echo "Gestor del Servidor MIA"
    echo ""
    echo "Uso: $0 {start|stop|restart|status|logs|build|help}"
    echo ""
    echo "Comandos:"
    echo "  start    - Iniciar el servidor"
    echo "  stop     - Detener el servidor"
    echo "  restart  - Reiniciar el servidor"
    echo "  status   - Ver estado del servidor"
    echo "  logs     - Ver últimas líneas de logs"
    echo "  logs -f  - Ver logs en tiempo real"
    echo "  build    - Compilar el servidor"
    echo "  help     - Mostrar esta ayuda"
    echo ""
    echo "Ejemplos:"
    echo "  $0 start         # Iniciar servidor"
    echo "  $0 logs -f       # Ver logs en tiempo real"
    echo "  $0 restart       # Reiniciar servidor"
}

# Comando principal
case "${1:-}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs "$2"
        ;;
    build)
        build
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Comando inválido: ${1:-}"
        echo ""
        show_help
        exit 1
        ;;
esac
