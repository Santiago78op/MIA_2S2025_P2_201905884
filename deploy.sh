#!/bin/bash

# Script de Despliegue Automatizado - MIA Proyecto 2
# Uso: ./deploy.sh [backend|frontend|all]

set -e  # Exit on error

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Funciones de logging
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Directorio raíz del proyecto
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_ROOT/Backend"
FRONTEND_DIR="$PROJECT_ROOT/Frontend"

# Variables de configuración (editar según tu setup)
EC2_USER="${EC2_USER:-ubuntu}"
EC2_HOST="${EC2_HOST}"
EC2_KEY="${EC2_KEY}"
S3_BUCKET="${S3_BUCKET}"
API_URL="${API_URL}"

# Función para verificar prerequisitos
check_prerequisites() {
    log_info "Verificando prerequisitos..."

    # Verificar Go
    if ! command -v go &> /dev/null; then
        log_error "Go no está instalado"
        exit 1
    fi
    log_info "✓ Go $(go version | awk '{print $3}')"

    # Verificar Node/npm
    if ! command -v npm &> /dev/null; then
        log_error "npm no está instalado"
        exit 1
    fi
    log_info "✓ npm $(npm --version)"

    log_info "Prerequisitos OK"
}

# Función para compilar backend
build_backend() {
    log_info "Compilando backend..."

    cd "$BACKEND_DIR"

    # Compilar para Linux (EC2)
    log_info "Compilando para Linux AMD64..."
    GOOS=linux GOARCH=amd64 go build -o bin/server-linux cmd/server/main.go

    if [ $? -eq 0 ]; then
        log_info "✓ Backend compilado: bin/server-linux"
        ls -lh bin/server-linux
    else
        log_error "Error compilando backend"
        exit 1
    fi
}

# Función para desplegar backend en EC2
deploy_backend() {
    log_info "Desplegando backend en EC2..."

    if [ -z "$EC2_HOST" ] || [ -z "$EC2_KEY" ]; then
        log_error "EC2_HOST y EC2_KEY deben estar configurados"
        log_info "Uso: EC2_HOST=1.2.3.4 EC2_KEY=/path/to/key.pem ./deploy.sh backend"
        exit 1
    fi

    # Verificar conexión SSH
    log_info "Verificando conexión con $EC2_USER@$EC2_HOST..."
    ssh -i "$EC2_KEY" -o ConnectTimeout=10 "$EC2_USER@$EC2_HOST" "echo 'Conexión OK'" || {
        log_error "No se pudo conectar a EC2"
        exit 1
    }

    # Crear directorios en EC2
    log_info "Creando directorios en EC2..."
    ssh -i "$EC2_KEY" "$EC2_USER@$EC2_HOST" "mkdir -p ~/mia-backend/Backend/bin ~/Discos ~/Reports"

    # Copiar binario
    log_info "Copiando binario..."
    scp -i "$EC2_KEY" "$BACKEND_DIR/bin/server-linux" "$EC2_USER@$EC2_HOST:~/mia-backend/Backend/bin/server"

    # Copiar directorios necesarios (si no existen)
    log_info "Copiando directorios de datos..."
    scp -r -i "$EC2_KEY" "$BACKEND_DIR/Discos" "$EC2_USER@$EC2_HOST:~/" 2>/dev/null || true
    scp -r -i "$EC2_KEY" "$BACKEND_DIR/Reports" "$EC2_USER@$EC2_HOST:~/" 2>/dev/null || true

    # Reiniciar servicio (si existe)
    log_info "Reiniciando servicio..."
    ssh -i "$EC2_KEY" "$EC2_USER@$EC2_HOST" "sudo systemctl restart mia-backend 2>/dev/null || echo 'Servicio no configurado (primera vez)'"

    # Verificar health
    log_info "Esperando 5 segundos para que el servicio inicie..."
    sleep 5

    log_info "Verificando health endpoint..."
    if curl -s "http://$EC2_HOST:8080/health" > /dev/null; then
        log_info "✓ Backend desplegado correctamente"
        log_info "URL: http://$EC2_HOST:8080"
    else
        log_warn "El endpoint /health no responde. Verifica los logs en EC2:"
        log_warn "  ssh -i $EC2_KEY $EC2_USER@$EC2_HOST 'tail -f ~/backend.log'"
    fi
}

# Función para compilar frontend
build_frontend() {
    log_info "Compilando frontend..."

    cd "$FRONTEND_DIR"

    # Crear .env.production si API_URL está configurado
    if [ -n "$API_URL" ]; then
        log_info "Configurando API_URL=$API_URL"
        echo "VITE_API_URL=$API_URL" > .env.production
    else
        log_warn "API_URL no configurado. El frontend usará proxy local."
    fi

    # Instalar dependencias (si es necesario)
    if [ ! -d "node_modules" ]; then
        log_info "Instalando dependencias npm..."
        npm install
    fi

    # Compilar
    log_info "Compilando con Vite..."
    npm run build

    if [ $? -eq 0 ]; then
        log_info "✓ Frontend compilado en dist/"
        du -sh dist
    else
        log_error "Error compilando frontend"
        exit 1
    fi
}

# Función para desplegar frontend en S3
deploy_frontend() {
    log_info "Desplegando frontend en S3..."

    if [ -z "$S3_BUCKET" ]; then
        log_error "S3_BUCKET debe estar configurado"
        log_info "Uso: S3_BUCKET=mi-bucket ./deploy.sh frontend"
        exit 1
    fi

    # Verificar AWS CLI
    if ! command -v aws &> /dev/null; then
        log_error "AWS CLI no está instalado"
        log_info "Instalar: https://aws.amazon.com/cli/"
        exit 1
    fi

    # Verificar credenciales
    if ! aws sts get-caller-identity &> /dev/null; then
        log_error "AWS CLI no está configurado. Ejecuta 'aws configure'"
        exit 1
    fi

    # Subir a S3
    log_info "Subiendo archivos a s3://$S3_BUCKET/..."
    aws s3 sync "$FRONTEND_DIR/dist/" "s3://$S3_BUCKET/" --delete --acl public-read

    if [ $? -eq 0 ]; then
        log_info "✓ Frontend desplegado correctamente"

        # Obtener URL del bucket
        REGION=$(aws s3api get-bucket-location --bucket "$S3_BUCKET" --query "LocationConstraint" --output text)
        if [ "$REGION" = "None" ] || [ -z "$REGION" ]; then
            REGION="us-east-1"
        fi

        S3_URL="http://$S3_BUCKET.s3-website-$REGION.amazonaws.com"
        log_info "URL: $S3_URL"

        log_warn "Si el bucket no tiene static website hosting habilitado, configúralo en AWS Console:"
        log_warn "  S3 → $S3_BUCKET → Properties → Static website hosting → Enable"
    else
        log_error "Error subiendo a S3"
        exit 1
    fi
}

# Función principal
main() {
    local command=${1:-all}

    echo "========================================="
    echo "  MIA Proyecto 2 - Deploy Script"
    echo "========================================="
    echo ""

    check_prerequisites

    case $command in
        backend)
            build_backend
            deploy_backend
            ;;
        frontend)
            build_frontend
            deploy_frontend
            ;;
        build-backend)
            build_backend
            ;;
        build-frontend)
            build_frontend
            ;;
        all)
            build_backend
            deploy_backend
            build_frontend
            deploy_frontend
            ;;
        *)
            log_error "Comando inválido: $command"
            echo ""
            echo "Uso: ./deploy.sh [comando]"
            echo ""
            echo "Comandos disponibles:"
            echo "  backend         - Compila y despliega backend en EC2"
            echo "  frontend        - Compila y despliega frontend en S3"
            echo "  build-backend   - Solo compila backend"
            echo "  build-frontend  - Solo compila frontend"
            echo "  all             - Despliega todo (default)"
            echo ""
            echo "Variables de entorno requeridas:"
            echo "  EC2_HOST    - IP pública de EC2"
            echo "  EC2_KEY     - Path a la key SSH (.pem)"
            echo "  EC2_USER    - Usuario SSH (default: ubuntu)"
            echo "  S3_BUCKET   - Nombre del bucket S3"
            echo "  API_URL     - URL del backend para el frontend"
            echo ""
            echo "Ejemplo:"
            echo "  EC2_HOST=1.2.3.4 EC2_KEY=~/.ssh/key.pem S3_BUCKET=my-bucket API_URL=http://1.2.3.4:8080 ./deploy.sh all"
            exit 1
            ;;
    esac

    echo ""
    log_info "¡Despliegue completado! 🚀"
}

# Ejecutar
main "$@"
