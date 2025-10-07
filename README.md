# GoDisk - Sistema de Gestión de Discos

## Descripción

GoDisk es un sistema completo de gestión y simulación de discos que permite crear, formatear, montar y administrar sistemas de archivos virtuales. El proyecto está desarrollado con una arquitectura de cliente-servidor, donde el backend está implementado en Go y el frontend en React con TypeScript.

## Características Principales

### Backend (Go)

- **Gestión de Discos**: Creación y administración de discos virtuales
- **Sistemas de Archivos**: Soporte para EXT2 y EXT3 con journaling
- **Particionamiento**: Manejo de MBR (Master Boot Record) y EBR (Extended Boot Record)
- **Reportes**: Generación de reportes en formato DOT para visualización
- **Journaling**: Sistema de journal para EXT3 con recuperación de transacciones
- **API REST**: Servidor HTTP con endpoints para todas las operaciones
- **Logging**: Sistema de logs detallado para auditoría

### Frontend (React + TypeScript)

- **Terminal de Comandos**: Interfaz interactiva para ejecutar comandos del sistema
- **Explorador de Discos**: Navegación visual del sistema de archivos
- **Visualizador de Reportes**: Renderizado de diagramas DOT
- **Panel de Journal**: Monitoreo de transacciones del sistema de archivos
- **Visor de Logs**: Consulta de logs del sistema en tiempo real
- **Ejecutor de Scripts**: Capacidad de ejecutar scripts de comandos

## Estructura del Proyecto

```text
├── Backend/
│   ├── cmd/server/          # Servidor HTTP y handlers
│   ├── internal/
│   │   ├── commands/        # Procesamiento de comandos
│   │   ├── disk/           # Gestión de discos y particiones
│   │   ├── fs/             # Sistemas de archivos (EXT2/EXT3)
│   │   └── journal/        # Sistema de journaling
│   ├── pkg/reports/        # Generación de reportes
│   └── scripts/            # Scripts de utilidad
└── Frontend/
    └── godisk-frontend/
        ├── src/
        │   ├── components/     # Componentes React
        │   ├── pages/         # Páginas principales
        │   ├── lib/           # API y utilidades
        │   └── types/         # Definiciones TypeScript
        └── public/            # Archivos estáticos
```

## Funcionalidades

### Comandos de Disco

- `mkdisk`: Crear discos virtuales
- `rmdisk`: Eliminar discos
- `fdisk`: Crear y administrar particiones
- `mount`: Montar particiones
- `unmount`: Desmontar particiones

### Comandos de Sistema de Archivos

- `mkfs`: Formatear particiones con EXT2 o EXT3
- `login`: Autenticación de usuarios
- `logout`: Cerrar sesión
- `mkgrp`: Crear grupos de usuarios
- `rmgrp`: Eliminar grupos
- `mkusr`: Crear usuarios
- `rmusr`: Eliminar usuarios

### Comandos de Archivos y Directorios

- `mkfile`: Crear archivos
- `cat`: Mostrar contenido de archivos
- `remove`: Eliminar archivos y directorios
- `edit`: Editar archivos
- `rename`: Renombrar archivos/directorios
- `mkdir`: Crear directorios
- `copy`: Copiar archivos
- `move`: Mover archivos
- `find`: Buscar archivos
- `chown`: Cambiar propietario
- `chgrp`: Cambiar grupo
- `chmod`: Cambiar permisos

### Comandos de Reportes

- `rep`: Generar reportes visuales (disk, inode, journaling, block, bm_inode, bm_block, tree, sb, file, ls)

## Tecnologías Utilizadas

### Tecnologías Backend

- **Go 1.21+**: Lenguaje principal
- **Gorilla Mux**: Enrutamiento HTTP
- **CORS**: Manejo de políticas de origen cruzado

### Tecnologías Frontend

- **React 18**: Framework de UI
- **TypeScript**: Tipado estático
- **Vite**: Herramienta de construcción
- **Tailwind CSS**: Framework de estilos
- **Vis.js**: Visualización de grafos DOT

## Instalación y Uso

### Prerrequisitos

- Go 1.21 o superior
- Node.js 18 o superior
- npm o yarn

### Inicio Rápido

El proyecto incluye scripts automatizados para facilitar el desarrollo:

#### Opción 1: Scripts de Conveniencia (Recomendado)

```bash
# Desde la raíz del proyecto
./start.sh    # Inicia frontend y backend automáticamente
./stop.sh     # Mata todos los procesos del proyecto
```

#### Opción 2: Scripts Completos

```bash
# Desde la raíz del proyecto
./Backend/scripts/start-project.sh    # Script completo de inicio
./Backend/scripts/kill-processes.sh   # Script completo de limpieza
```

Los scripts automáticamente:

- ✅ Verifican dependencias (Go, Node.js, npm)
- ✅ Instalan paquetes si es necesario
- ✅ Limpian puertos en uso
- ✅ Inician ambos servicios en paralelo
- ✅ Muestran logs en tiempo real
- ✅ Permiten detener con Ctrl+C

### Configuración Manual

#### Configuración Backend

```bash
cd Backend
go mod download
go run cmd/server/main.go
```

#### Configuración Frontend

```bash
cd Frontend/godisk-frontend
npm install
npm run dev
```

### URLs de Acceso

Una vez iniciados los servicios:

- **Backend API**: <http://localhost:8080>
- **Frontend Web**: <http://localhost:5173>

## Contribución

Este proyecto es parte del curso de Manejo e Implementación de Archivos (MIA) y está en desarrollo activo.

## Licencia

Proyecto académico - Universidad de San Carlos de Guatemala
