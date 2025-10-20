#!/bin/bash

# Script de instalación de dependencias para MIA Proyecto 2
# Sistema Operativo: Ubuntu/Debian
# Autor: Generado automáticamente
# Fecha: 2025

set -e  # Detener en caso de error

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║   MIA Proyecto 2 - Instalación de Dependencias           ║"
echo "║   Sistema de Archivos EXT2/EXT3                          ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Función para imprimir mensajes
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# Verificar que se está ejecutando en Ubuntu/Debian
if ! [ -f /etc/os-release ]; then
    print_error "No se puede detectar el sistema operativo"
    exit 1
fi

source /etc/os-release
if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
    print_warning "Este script está diseñado para Ubuntu/Debian. Tu sistema es: $ID"
    read -p "¿Deseas continuar de todas formas? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

print_info "Sistema detectado: $PRETTY_NAME"
echo

# ============================================
# 1. Actualizar sistema
# ============================================
print_info "Actualizando lista de paquetes..."
sudo apt update

echo
read -p "¿Deseas actualizar el sistema completo? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    print_info "Actualizando sistema..."
    sudo apt upgrade -y
    print_success "Sistema actualizado"
fi

echo

# ============================================
# 2. Instalar dependencias básicas
# ============================================
print_info "Instalando dependencias básicas..."
sudo apt install -y \
    curl \
    wget \
    git \
    build-essential \
    software-properties-common \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release

print_success "Dependencias básicas instaladas"
echo

# ============================================
# 3. Instalar Go
# ============================================
GO_VERSION="1.23.3"
print_info "Verificando instalación de Go..."

if command -v go &> /dev/null; then
    CURRENT_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go ya está instalado: v$CURRENT_GO_VERSION"

    read -p "¿Deseas reinstalar Go v$GO_VERSION? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Manteniendo versión actual de Go"
    else
        INSTALL_GO=true
    fi
else
    print_warning "Go no está instalado"
    INSTALL_GO=true
fi

if [ "$INSTALL_GO" = true ]; then
    print_info "Descargando Go v$GO_VERSION..."

    # Detectar arquitectura
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
        GO_ARCH="amd64"
    elif [ "$ARCH" = "aarch64" ]; then
        GO_ARCH="arm64"
    else
        print_error "Arquitectura no soportada: $ARCH"
        exit 1
    fi

    GO_TARBALL="go${GO_VERSION}.linux-${GO_ARCH}.tar.gz"

    cd /tmp
    wget -q --show-progress "https://go.dev/dl/${GO_TARBALL}"

    print_info "Instalando Go..."
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "${GO_TARBALL}"
    rm "${GO_TARBALL}"

    # Configurar PATH si no está configurado
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo "" >> ~/.bashrc
        echo "# Go environment" >> ~/.bashrc
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
    fi

    export PATH=$PATH:/usr/local/go/bin

    print_success "Go v$GO_VERSION instalado correctamente"
    go version
fi

echo

# ============================================
# 4. Instalar Node.js y npm
# ============================================
NODE_VERSION="20"
print_info "Verificando instalación de Node.js..."

if command -v node &> /dev/null; then
    CURRENT_NODE_VERSION=$(node -v)
    print_success "Node.js ya está instalado: $CURRENT_NODE_VERSION"

    read -p "¿Deseas reinstalar Node.js v$NODE_VERSION? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Manteniendo versión actual de Node.js"
    else
        INSTALL_NODE=true
    fi
else
    print_warning "Node.js no está instalado"
    INSTALL_NODE=true
fi

if [ "$INSTALL_NODE" = true ]; then
    print_info "Instalando Node.js v$NODE_VERSION desde NodeSource..."

    # Agregar repositorio de NodeSource
    curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | sudo -E bash -

    # Instalar Node.js
    sudo apt install -y nodejs

    print_success "Node.js instalado correctamente"
    node -v
    npm -v
fi

echo

# ============================================
# 5. Instalar dependencias del Backend (Go)
# ============================================
print_info "Instalando dependencias del Backend (Go)..."

BACKEND_DIR="$(dirname "$0")/Backend"

if [ -d "$BACKEND_DIR" ]; then
    cd "$BACKEND_DIR"

    if [ -f "go.mod" ]; then
        print_info "Descargando módulos de Go..."
        go mod download
        go mod tidy
        print_success "Dependencias de Go instaladas"
    else
        print_warning "No se encontró go.mod en el directorio Backend"
    fi
else
    print_warning "No se encontró el directorio Backend"
fi

echo

# ============================================
# 6. Instalar dependencias del Frontend (npm)
# ============================================
print_info "Instalando dependencias del Frontend (npm)..."

FRONTEND_DIR="$(dirname "$0")/Frontend"

if [ -d "$FRONTEND_DIR" ]; then
    cd "$FRONTEND_DIR"

    if [ -f "package.json" ]; then
        print_info "Instalando paquetes npm..."
        npm install
        print_success "Dependencias de npm instaladas"
    else
        print_warning "No se encontró package.json en el directorio Frontend"
    fi
else
    print_warning "No se encontró el directorio Frontend"
fi

echo

# ============================================
# 7. Compilar Backend
# ============================================
print_info "Compilando Backend..."

cd "$BACKEND_DIR"

if [ -f "cmd/server/main.go" ]; then
    print_info "Compilando servidor Go..."
    mkdir -p bin
    go build -o bin/server cmd/server/main.go

    if [ -f "bin/server" ]; then
        print_success "Backend compilado exitosamente: bin/server"
    else
        print_error "Error al compilar el backend"
        exit 1
    fi
else
    print_warning "No se encontró cmd/server/main.go"
fi

echo

# ============================================
# 8. Crear directorios necesarios
# ============================================
print_info "Creando directorios necesarios..."

cd "$(dirname "$0")/Backend"

mkdir -p Discos
mkdir -p Reportes
mkdir -p logs

print_success "Directorios creados"
echo

# ============================================
# 9. Verificar instalación
# ============================================
print_info "Verificando instalación..."
echo

echo "├─ Sistema Operativo:"
echo "│  └─ $PRETTY_NAME"
echo "│"

if command -v go &> /dev/null; then
    echo "├─ Go:"
    echo "│  └─ $(go version)"
else
    echo "├─ Go: ${RED}No instalado${NC}"
fi
echo "│"

if command -v node &> /dev/null; then
    echo "├─ Node.js:"
    echo "│  └─ $(node -v)"
else
    echo "├─ Node.js: ${RED}No instalado${NC}"
fi
echo "│"

if command -v npm &> /dev/null; then
    echo "├─ npm:"
    echo "│  └─ $(npm -v)"
else
    echo "├─ npm: ${RED}No instalado${NC}"
fi
echo "│"

if [ -f "$BACKEND_DIR/bin/server" ]; then
    echo "├─ Backend:"
    echo "│  └─ ${GREEN}Compilado${NC} (bin/server)"
else
    echo "├─ Backend: ${RED}No compilado${NC}"
fi
echo "│"

if [ -d "$FRONTEND_DIR/node_modules" ]; then
    echo "└─ Frontend:"
    echo "   └─ ${GREEN}Dependencias instaladas${NC}"
else
    echo "└─ Frontend: ${RED}Dependencias no instaladas${NC}"
fi

echo
echo -e "${GREEN}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║   ✓ Instalación completada                               ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

echo
print_info "Para iniciar el proyecto:"
echo
echo "  Backend:"
echo "    cd Backend"
echo "    ./bin/server"
echo
echo "  Frontend (en otra terminal):"
echo "    cd Frontend"
echo "    npm run dev"
echo
echo "  Acceder a la aplicación:"
echo "    http://localhost:5173"
echo

print_info "Recuerda cerrar y reabrir la terminal para que los cambios de PATH tomen efecto"
echo

print_success "¡Todo listo! 🚀"
